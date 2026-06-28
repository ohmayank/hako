.PHONY: run test race vet fmt-check check build clean

ARGS ?= help

run:
	go run ./cmd/hako $(ARGS)

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

fmt-check:
	test -z "$$(gofmt -l .)"

check: fmt-check vet test race

build:
	mkdir -p bin
	go build -o bin/hako ./cmd/hako

clean:
	rm -rf bin
