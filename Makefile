APP_NAME = fartkeyboard
VERSION = 1.0.0
BUILD_DIR = build
MACOS_APP_DIR = $(BUILD_DIR)/$(APP_NAME).app
MACOS_EXECUTABLE = $(MACOS_APP_DIR)/Contents/MacOS/$(APP_NAME)

.PHONY: all clean build-macos package-macos

all: clean build-macos package-macos

build-macos:
	mkdir -p $(MACOS_APP_DIR)/Contents/MacOS
	mkdir -p $(MACOS_APP_DIR)/Contents/Resources
	echo '<?xml version="1.0" encoding="UTF-8"?>' > $(MACOS_APP_DIR)/Contents/Info.plist
	echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> $(MACOS_APP_DIR)/Contents/Info.plist
	echo '<plist version="1.0">' >> $(MACOS_APP_DIR)/Contents/Info.plist
	echo '<dict>' >> $(MACOS_APP_DIR)/Contents/Info.plist
	echo '  <key>CFBundleExecutable</key>' >> $(MACOS_APP_DIR)/Contents/Info.plist
	echo '  <string>$(APP_NAME)</string>' >> $(MACOS_APP_DIR)/Contents/Info.plist
	echo '  <key>CFBundleIdentifier</key>' >> $(MACOS_APP_DIR)/Contents/Info.plist
	echo '  <string>com.example.$(APP_NAME)</string>' >> $(MACOS_APP_DIR)/Contents/Info.plist
	echo '  <key>CFBundleName</key>' >> $(MACOS_APP_DIR)/Contents/Info.plist
	echo '  <string>$(APP_NAME)</string>' >> $(MACOS_APP_DIR)/Contents/Info.plist
	echo '  <key>CFBundleVersion</key>' >> $(MACOS_APP_DIR)/Contents/Info.plist
	echo '  <string>$(VERSION)</string>' >> $(MACOS_APP_DIR)/Contents/Info.plist
	echo '</dict>' >> $(MACOS_APP_DIR)/Contents/Info.plist
	echo '</plist>' >> $(MACOS_APP_DIR)/Contents/Info.plist
	GOOS=darwin GOARCH=arm64 go build -o $(MACOS_EXECUTABLE)

package-macos:
	cd $(BUILD_DIR) && zip -r $(APP_NAME)-macos-$(VERSION).zip $(APP_NAME).app

clean:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)
