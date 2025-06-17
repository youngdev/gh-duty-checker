BINARY_NAME=gh-duty-checker
SRC=main.go
VERSION=v1.0.0

build:
	@echo "Building for $(GOOS)/$(GOARCH)..."
	@go build -o bin/$(BINARY_NAME) $(SRC)

build-all: clean darwin-amd64 darwin-arm64 windows-amd64

darwin-amd64:
	@echo "Building for Darwin (AMD64)..."
	@GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-darwin-amd64-$(VERSION) $(SRC)

darwin-arm64:
	@echo "Building for Darwin (ARM64)..."
	@GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)-darwin-arm64-$(VERSION) $(SRC)

windows-amd64:
	@echo "Building for Windows (AMD64)..."
	@GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)-windows-amd64-$(VERSION).exe $(SRC)

clean:
	@echo "Cleaning up old builds..."
	@rm -f bin/$(BINARY_NAME)*

install:
	@echo "Installing gh-duty-checker..."
	@go install

.PHONY: build build-all clean install darwin-amd64 darwin-arm64 windows-amd64 