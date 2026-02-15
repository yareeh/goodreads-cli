# goodreads-cli

A command-line interface for Goodreads. Uses browser automation (rod) for login and shelf operations, and plain HTTP for search.

## Setup

1. Install [Go](https://go.dev/dl/)

2. Build:
   ```
   go build -o goodreads .
   ```

3. Create `~/.goodreads-cli.yaml` with your Goodreads (Amazon) credentials:
   ```yaml
   email: you@example.com
   password: yourpassword
   ```

## Usage

### Login

```
./goodreads login
```

Launches a headless browser, navigates through Amazon's OpenID login flow, and saves the session to `~/.goodreads-cli-session`.

### Search

```
./goodreads search "project hail mary"
```

Searches Goodreads and displays results as a table with book IDs, titles, and authors. Works without login.

### Add to shelf

```
./goodreads shelf 55145261 --shelf want-to-read
./goodreads shelf 55145261 --shelf currently-reading
./goodreads shelf 55145261 --shelf read
```

### Start reading a book

```
./goodreads new 55145261
```

Shortcut for adding a book to the `currently-reading` shelf.

### Finish a book

```
./goodreads finished 55145261
```

Shortcut for adding a book to the `read` shelf.

### Reply to a discussion topic

```
./goodreads post-reply 1585066 --message "Hello from the CLI!"
./goodreads post-reply 1585066 --message "Check out this book" --book 55145261
./goodreads post-reply 1585066 --message "Great author" --author 513351
```

Posts a reply to an existing discussion topic. The argument is the topic ID from the URL (e.g. `goodreads.com/topic/show/1585066`). Use `--book` or `--author` to add a reference link.

### Create a new discussion topic

```
./goodreads post-topic \
  --url "https://www.goodreads.com/topic/new?context_id=220-group&context_type=Group&topic[folder_id]=120471" \
  --subject "My topic title" \
  --message "The body of the topic"
./goodreads post-topic \
  --url "https://www.goodreads.com/topic/new?context_id=220-group&context_type=Group&topic[folder_id]=120471" \
  --subject "Book recommendation" \
  --message "You should read this" \
  --book 55145261
```

Creates a new topic in a group. The `--url` is the full new-topic URL from Goodreads (copy it from the "New topic" link in the group). Use `--book` or `--author` to add a reference link.

## Debugging

Add `--no-headless` to any command to open a visible browser window:

```
./goodreads login --no-headless
./goodreads shelf 55145261 --shelf want-to-read --no-headless
./goodreads post-reply 1585066 --message "test" --no-headless
```

This is useful for diagnosing issues (CAPTCHAs, 2FA prompts, changed page layouts). When any browser command fails, a debug screenshot is saved to `~/goodreads-cli-debug.png`.

## AI Agent Integration

This CLI is designed to be easily scriptable and can be used as a tool/skill by AI agents and automation frameworks. See [SKILL.md](SKILL.md) for the full agent reference including command documentation and common workflows like searching for a book by name and adding it to a shelf.

The CLI uses plain text output and standard exit codes, making it straightforward to integrate with any agent framework, shell script, or automation tool.

**Claude Code:** Add to `.claude/settings.json`:
```json
{ "skills": ["/path/to/goodreads-cli/SKILL.md"] }
```

**Other agents:** Point your agent's tool/skill config at `SKILL.md`, or include its contents in your system prompt.

## How it works

- **Search** uses Goodreads' JSON autocomplete endpoint (`/book/auto_complete?format=json`) via plain HTTP
- **Login** and **shelf operations** use [rod](https://github.com/go-rod/rod) for headless browser automation, since Goodreads routes login through Amazon's OpenID and shelf mutations go through Next.js/React internals
- Session cookies are persisted to `~/.goodreads-cli-session` so you only need to log in once
