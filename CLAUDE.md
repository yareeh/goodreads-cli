# goodreads-cli

Go CLI for Goodreads using browser automation (rod) and HTTP scraping.

## Build

```
go build -o goodreads .
```

## Releasing

Tag and release using `gh`:

```bash
# Tag the release
git tag v0.1.0
git push origin v0.1.0

# Create GitHub release
gh release create v0.1.0 --generate-notes

# Or with a title and notes
gh release create v0.1.0 --title "v0.1.0" --notes "First release"
```

To attach binaries, cross-compile first:

```bash
GOOS=linux GOARCH=amd64 go build -o goodreads-cli-linux-amd64 .
GOOS=darwin GOARCH=arm64 go build -o goodreads-cli-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o goodreads-cli-windows-amd64.exe .
gh release create v0.1.0 --generate-notes \
  goodreads-cli-linux-amd64 \
  goodreads-cli-darwin-arm64 \
  goodreads-cli-windows-amd64.exe
```

## Codecov

Coverage is uploaded automatically by CI. To activate:
1. Sign in to [codecov.io](https://codecov.io) with GitHub
2. Add the repo
3. Add `CODECOV_TOKEN` to repo secrets (Settings > Secrets)
