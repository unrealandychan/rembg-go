.PHONY: build tidy test

build:
	go build -o bin/rembg ./cmd/rembg

tidy:
	go mod tidy

test:
	go test ./...
