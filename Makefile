.PHONY: build

build:
	make clean
	go build -o build/auto-storage .

build-linux:
	make clean
	GOOS=linux GOARCH=amd64 go build -o build/auto-storage-linux-amd64 ./src

install:
	make build
	sudo cp build/auto-storage /usr/local/bin

deps:
	go mod tidy

clean:
	rm -rf build
	go clean
