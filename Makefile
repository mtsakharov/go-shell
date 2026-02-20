BINARY  := gosh
PKG     := ./cmd/gosh
GOFLAGS := -trimpath

.PHONY: build run clean test lint

build:
	go build $(GOFLAGS) -o $(BINARY) $(PKG)

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)

test:
	go test ./...

lint:
	go vet ./...
