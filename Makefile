.PHONY: all build run clean

all: build

build:
	go build -o koi .

run: build
	./koi -web ./web

clean:
	rm -f koi
	rm -rf data/
