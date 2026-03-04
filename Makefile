BINARY := prr

.PHONY: build test clean

build:
	go build -o ./$(BINARY) ./cmd/prr

test:
	go test ./...

clean:
	rm -f ./$(BINARY)
