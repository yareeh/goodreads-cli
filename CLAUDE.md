# goodreads-cli

Go CLI for Goodreads using browser automation (rod) and HTTP scraping.

## Build

```
go build -o goodreads .
```

## Test

Fast loop (unit tests + parser fixture, no network, no browser):

```
go test ./...
```

Full suite including live-Goodreads integration tests (browser via rod,
~100s, sensitive to Goodreads rate limits — set GOODREADS_EMAIL and
GOODREADS_PASSWORD or use .env):

```
go test -tags integration ./...
```

If logins fail with "login failed" and the debug HTML shows an Amazon
CAPTCHA (`amzn-captcha-root`), the account is being bot-checked. Run a headed
login from the repo root, solve the CAPTCHA in the window, and the tests will
reuse the saved session. Tests and this helper use their own session store,
`./.goodreads-cli-session` (gitignored, via `GOODREADS_SESSION_PATH`), never
the CLI's `~/.goodreads-cli-session`:

```
go run ./cmd/headed-login
```

**The gated suite must be run and green before any `git push origin`.**
The build tag exists to keep iteration fast, not to let slow tests rot.


## Development workflow (TDD)

For every new feature or bug fix:
1. **Write tests first**
2. **Implement** the feature
3. **Lint and format**
4. **Run tests**
5. Fix any issues and repeat until clean

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
