#!/usr/bin/env bash
# Run every static check CI's "test" job runs (no network tests, no browser),
# so problems surface before a push instead of in CI. Used by the pre-push
# hook (scripts/hooks/pre-push); run it by hand with `make static`.
#
# Mirrors .github/workflows/ci.yml: pre-commit hooks (gitleaks, gofmt, vet),
# go mod verify, gosec and govulncheck (both @latest, as CI does).
#
# The Go toolchain is pinned to go.mod's `go` line, the version CI's setup-go
# uses, so govulncheck scans the same standard library CI does rather than
# whatever newer Go happens to be installed locally. The scanners run via
# `go run ...@latest` under that toolchain, so a stale binary in ~/go/bin
# built with another Go version cannot break them.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

GOTOOLCHAIN="go$(go mod edit -json | sed -n 's/^[[:space:]]*"Go": "\(.*\)",*$/\1/p')"
export GOTOOLCHAIN

step() { printf '\n==> %s\n' "$*"; }

step "toolchain $(go version)"

if command -v pre-commit >/dev/null 2>&1; then
	step "pre-commit (gitleaks, gofmt, go vet)"
	pre-commit run --all-files
else
	step "gofmt"
	unformatted="$(gofmt -l $(git ls-files '*.go'))"
	if [ -n "$unformatted" ]; then
		echo "gofmt needed on:"
		echo "$unformatted"
		exit 1
	fi
	step "go vet"
	go vet ./...
fi

# Secret scan of the commits about to leave this machine (reachable from HEAD,
# on no remote yet). Same gitleaks version as .pre-commit-config.yaml, whose
# hook only sees staged changes.
step "gitleaks (unpushed commits)"
go run github.com/zricethezav/gitleaks/v8@v8.18.4 detect --no-banner --redact --source . \
	--log-opts="HEAD --not --remotes"

step "go vet (integration tag)"
go vet -tags integration ./...

step "go mod verify"
go mod verify

step "gosec"
go run github.com/securego/gosec/v2/cmd/gosec@latest -quiet ./...

step "govulncheck"
go run golang.org/x/vuln/cmd/govulncheck@latest ./...

step "all static checks passed"
