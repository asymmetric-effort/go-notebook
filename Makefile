.PHONY: all clean lint test build release hooks

-include Makefile.d/*.mk

all: lint test build

clean:
	go run ./tools/clean

lint:
	go run ./tools/lint

test:
	go run ./tools/test

build:
	go run ./tools/build

release:
	go run ./tools/release

hooks:
	go run ./tools/hooks
