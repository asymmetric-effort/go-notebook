.PHONY: all clean lint test build release hooks version version/major version/minor version/patch editor

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

version:
	go run ./tools/version

version/major:
	go run ./tools/version major

version/minor:
	go run ./tools/version minor

version/patch:
	go run ./tools/version patch

editor:
	cd nb/server/editor && bun install --frozen-lockfile && bun run build
