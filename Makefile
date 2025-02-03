BINARY_NAME=cryptonite
VERSION=0.0.1

# Build directories
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin

# Platforms
PLATFORMS=linux darwin windows
ARCHITECTURES=amd64 arm64

.PHONY: all clean install uninstall

all: clean build

build:
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			if [ "$$platform" = "windows" ]; then \
				GOOS=$$platform GOARCH=$$arch go build -o $(BUILD_DIR)/$(BINARY_NAME)-$$platform-$$arch.exe ./cmd/main.go; \
			else \
				GOOS=$$platform GOARCH=$$arch go build -o $(BUILD_DIR)/$(BINARY_NAME)-$$platform-$$arch ./cmd/main.go; \
			fi; \
		done; \
	done

clean:
	@rm -rf $(BUILD_DIR)

# Install for current platform
install:
	@if [ "$$(uname)" = "Darwin" ]; then \
		cp $(BUILD_DIR)/$(BINARY_NAME)-darwin-$$(uname -m) $(INSTALL_DIR)/$(BINARY_NAME); \
	elif [ "$$(uname)" = "Linux" ]; then \
		cp $(BUILD_DIR)/$(BINARY_NAME)-linux-$$(uname -m) $(INSTALL_DIR)/$(BINARY_NAME); \
	else \
		echo "Unsupported platform"; \
		exit 1; \
	fi
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)

uninstall:
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME) 