.PHONY: build

build:
	make clean
	go build -o build/auto-storage ./src

install:
	make build
	sudo cp build/auto-storage /usr/local/bin

deps:
	go mod tidy

clean:
	rm -rf build
	go clean
