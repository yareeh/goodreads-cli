.PHONY: all setup build test lint clean

all: build

setup:
	@command -v pre-commit >/dev/null 2>&1 || { echo "pre-commit not found. Install with: pip install pre-commit"; exit 1; }
	pre-commit install

build:
	go build -mod=readonly -o goodreads .
	go build -mod=readonly -o goodreads-recorder ./cmd/recorder

test:
	go test -mod=readonly -v -short ./...

lint:
	go vet ./...
	pre-commit run --all-files

clean:
	rm -f goodreads goodreads-recorder
