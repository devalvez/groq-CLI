.PHONY: build install clean run test

BINARY := groq
BUILD_FLAGS := -ldflags="-s -w"
INSTALL_DIR := /usr/local/bin

build:
	@echo "⚡ Building groq-cli..."
	@go build $(BUILD_FLAGS) -o $(BINARY) .
	@echo "✅ Built: ./$(BINARY)"

install: build
	@echo "📦 Installing to $(INSTALL_DIR)..."
	@sudo mv $(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "✅ Installed: $(INSTALL_DIR)/$(BINARY)"

clean:
	@rm -f $(BINARY)
	@echo "🧹 Cleaned"

run:
	@go run . $(ARGS)

deps:
	@go mod tidy
	@echo "✅ Dependencies updated"

test:
	@go test ./...

# Cross-platform builds
build-all:
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o dist/groq-linux-amd64 .
	@GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o dist/groq-linux-arm64 .
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o dist/groq-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o dist/groq-darwin-arm64 .
	@echo "✅ All builds complete"

.DEFAULT_GOAL := build
