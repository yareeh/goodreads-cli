.PHONY: all setup hooks static build test lint clean

all: build

setup: hooks
	@command -v pre-commit >/dev/null 2>&1 || { echo "pre-commit not found. Install with: pip install pre-commit"; exit 1; }
	pre-commit install

# Install the repo's git hooks (pre-push runs `make static`) into this clone.
hooks:
	scripts/install-hooks.sh

# Every static check CI runs: pre-commit hooks, go mod verify, gosec, govulncheck.
static:
	scripts/static-checks.sh

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
