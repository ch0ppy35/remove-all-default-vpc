BIN_NAME=remove-all-default-vpc

PHONY: build clean

build: clean
	CGO_ENABLED=0 go build -ldflags='-s -w' -o bin/${BIN_NAME} .

clean:
	rm -rf bin/${BIN_NAME}
