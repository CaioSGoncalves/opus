.DEFAULT_GOAL := build
.PHONY: fmt vet test clean build run install

TARGET := opus

fmt:
	go fmt ./cmd/$(TARGET)

vet:
	go vet ./cmd/$(TARGET)

test:
	go test ./cmd/$(TARGET)

clean:
	rm -rf bin/

build: fmt vet test
	go build -o bin/ ./cmd/$(TARGET)

run: fmt vet test
	go run ./cmd/$(TARGET)

install: fmt vet test
	go install ./cmd/$(TARGET)
