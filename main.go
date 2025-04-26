package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	hook "github.com/robotn/gohook"
)

//go:embed mp3 ogg
var samplesFS embed.FS

const (
	historyLen      = 10
	throttlePeriod  = 700 * time.Millisecond
	resampleQuality = 3
)

var targetSampleRate = beep.SampleRate(48000)

func loadBuffers() ([]*beep.Buffer, error) {
	var buffers []*beep.Buffer

	dirs := []string{"mp3", "ogg"}
	for _, dir := range dirs {
		entries, err := fs.ReadDir(samplesFS, dir)
		if err != nil {
			return nil, fmt.Errorf("reading directory %q: %w", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			path := filepath.Join(dir, entry.Name())
			f, err := samplesFS.Open(path)
			if err != nil {
				return nil, fmt.Errorf("opening file %q: %w", path, err)
			}

			streamSeek, format, err := decodeAudio(f, entry.Name())
			f.Close()
			if err != nil {
				return nil, fmt.Errorf("decoding file %q: %w", path, err)
			}

			var streamer beep.Streamer = streamSeek
			if format.SampleRate != targetSampleRate {
				streamer = beep.Resample(resampleQuality, format.SampleRate, targetSampleRate, streamer)
			}

			buf := beep.NewBuffer(beep.Format{
				SampleRate:  targetSampleRate,
				NumChannels: format.NumChannels,
				Precision:   format.Precision,
			})
			buf.Append(streamer)
			streamSeek.Close()

			buffers = append(buffers, buf)
		}
	}

	if len(buffers) == 0 {
		return nil, fmt.Errorf("no samples found")
	}
	return buffers, nil
}

func decodeAudio(f fs.File, name string) (beep.StreamSeekCloser, beep.Format, error) {
	switch filepath.Ext(name) {
	case ".mp3":
		return mp3.Decode(f)
	case ".ogg":
		return vorbis.Decode(f)
	default:
		return nil, beep.Format{}, fmt.Errorf("unsupported file type: %s", name)
	}
}

func calculateWeights(buffers []*beep.Buffer) []float64 {
	weights := make([]float64, len(buffers))
	for i, buf := range buffers {
		duration := float64(buf.Len()) / float64(targetSampleRate)
		if duration > 0 {
			weights[i] = 1.0 / duration
		}
	}
	return weights
}

func chooseWeighted(weights []float64) int {
	total := 0.0
	for _, w := range weights {
		total += w
	}
	r := rand.Float64() * total
	for i, w := range weights {
		r -= w
		if r <= 0 {
			return i
		}
	}
	return len(weights) - 1
}

func sum(weights []float64) float64 {
	total := 0.0
	for _, w := range weights {
		total += w
	}
	return total
}

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Fart keyboard activated. Press Ctrl+C to quit.")

	buffers, err := loadBuffers()
	if err != nil {
		log.Fatalf("✗ failed to load samples: %v", err)
	}

	if err := speaker.Init(targetSampleRate, targetSampleRate.N(time.Millisecond*100)); err != nil {
		log.Fatalf("✗ failed to initialize speaker: %v", err)
	}

	mixer := &beep.Mixer{}
	speaker.Play(mixer)

	baseWeights := calculateWeights(buffers)
	history := make([]int, 0, historyLen)
	lastPlay := time.Now().Add(-throttlePeriod)

	events := hook.Start()
	defer hook.End()

	for ev := range events {
		if ev.Kind != hook.KeyDown {
			continue
		}
		if time.Since(lastPlay) < throttlePeriod {
			continue
		}

		weights := make([]float64, len(baseWeights))
		copy(weights, baseWeights)
		for _, idx := range history {
			if idx >= 0 && idx < len(weights) {
				weights[idx] = 0
			}
		}

		var idx int
		if sum(weights) > 0 {
			idx = chooseWeighted(weights)
		} else {
			idx = chooseWeighted(baseWeights)
		}

		streamer := buffers[idx].Streamer(0, buffers[idx].Len())
		mixer.Add(streamer)
		lastPlay = time.Now()

		history = append(history, idx)
		if len(history) > historyLen {
			history = history[1:]
		}
	}
}
