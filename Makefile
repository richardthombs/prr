BINARY := prr

# Source-first workflow.
# Canonical cross-platform commands are: go build ./... and go test ./...

.PHONY: build install test clean

build:
	go build -o ./$(BINARY) ./cmd/prr

install: build
	mkdir -p ~/.local/bin
	cp -f ./$(BINARY) ~/.local/bin/$(BINARY)

test:
	go test ./...

clean:
	rm -f ./$(BINARY)
