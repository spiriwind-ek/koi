.PHONY: all build run clean

all: build

build:
	CGO_ENABLED=1 go build -o koi .

run: build
	./koi

clean:
	rm -f koi
	rm -rf data/
