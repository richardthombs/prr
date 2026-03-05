BINARY := prr
BUILD_DIR := dist
GOARCH ?= amd64

.PHONY: build build-darwin build-linux build-windows build-all test clean

build:
	go build -o ./$(BINARY) ./cmd/prr

build-darwin:
	mkdir -p ./$(BUILD_DIR)
	GOOS=darwin GOARCH=$(GOARCH) go build -o ./$(BUILD_DIR)/$(BINARY)-darwin-$(GOARCH) ./cmd/prr

build-linux:
	mkdir -p ./$(BUILD_DIR)
	GOOS=linux GOARCH=$(GOARCH) go build -o ./$(BUILD_DIR)/$(BINARY)-linux-$(GOARCH) ./cmd/prr

build-windows:
	mkdir -p ./$(BUILD_DIR)
	GOOS=windows GOARCH=$(GOARCH) go build -o ./$(BUILD_DIR)/$(BINARY)-windows-$(GOARCH).exe ./cmd/prr

build-all: build-darwin build-linux build-windows

test:
	go test ./...

clean:
	rm -f ./$(BINARY)
	rm -rf ./$(BUILD_DIR)
