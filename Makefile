.PHONY: run test build clean

run:
	go run .

test:
	go test ./...

build:
	mkdir -p bin
	go build -o bin/hako .

clean:
	rm -rf bin

