.DEFAULT_GOAL := build
.PHONY: fmt vet test clean build run install

fmt:
	go fmt ./cmd/homelab

vet:
	go vet ./cmd/homelab

test:
	go test ./cmd/homelab

clean:
	rm -rf bin/

build: fmt vet test
	go build -o bin/ ./cmd/homelab

run: fmt vet test
	go run ./cmd/homelab

install: fmt vet test
	go install ./cmd/homelab
