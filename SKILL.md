# Goodreads CLI Skill

A headless browser-based CLI for interacting with Goodreads. Allows searching books, managing shelves, and posting to discussions.

## Prerequisites

- Go installed
- Chrome/Chromium available (rod downloads it automatically if missing)
- Config file at `~/.goodreads-cli.yaml`:
  ```yaml
  email: user@example.com
  password: theirpassword
  ```

## Build

```bash
cd /path/to/goodreads-cli
go build -o goodreads .
```

## Authentication

Login must be done before any shelf or discussion operations:

```bash
./goodreads login
```

Session is saved to `~/.goodreads-cli-session` and reused across commands. Login only needs to be done once (or when the session expires).

To check if login works, try a shelf operation — it will error with "not logged in" if the session is invalid.

## Commands

### Search for a book

```bash
./goodreads search "book title or author"
```

Returns a table with columns: ID, TITLE, AUTHOR. The ID is needed for all other book operations. **Does not require login.**

Example output:
```
ID           TITLE                                              AUTHOR
---          -----                                              ------
228233676    Rikkomuksia                                        Louise Kennedy
```

### Add a book to a shelf

```bash
./goodreads shelf <book-id> --shelf <shelf-name>
```

Shelf names: `want-to-read`, `currently-reading`, `read`

### Mark a book as currently reading

```bash
./goodreads new <book-id>
```

Shortcut for `shelf <book-id> --shelf currently-reading`. Works on both unshelved books and books already on another shelf.

### Mark a book as finished

```bash
./goodreads finished <book-id>
```

Shortcut for `shelf <book-id> --shelf read`.

### Reply to a discussion topic

```bash
./goodreads post-reply <topic-id> --message "text"
./goodreads post-reply <topic-id> --message "text" --book <book-id>
./goodreads post-reply <topic-id> --message "text" --author <author-id>
```

The topic ID is the number from the URL `goodreads.com/topic/show/<topic-id>`.

### Create a new discussion topic

```bash
./goodreads post-topic --url "<new-topic-url>" --subject "Title" --message "Body"
./goodreads post-topic --url "<new-topic-url>" --subject "Title" --message "Body" --book <book-id>
```

The `--url` is the full new-topic URL copied from Goodreads (includes group context and folder ID).

### Debugging

Add `--no-headless` to any command to show the browser window. On failure, a screenshot is saved to `~/goodreads-cli-debug.png`.

## Common Agent Workflows

### Find a book and add it to a shelf by name

1. Search for the book:
   ```bash
   ./goodreads search "Project Hail Mary"
   ```
2. Parse the book ID from the output (first column)
3. Add to shelf:
   ```bash
   ./goodreads shelf 55145261 --shelf want-to-read
   ```

### Start reading a book by name

1. Search: `./goodreads search "Rikkomuksia"`
2. Get the ID from output: `228233676`
3. Mark as reading: `./goodreads new 228233676`

### Find a book link for a user

Search returns book IDs. The Goodreads URL for any book is:
```
https://www.goodreads.com/book/show/<book-id>
```

Author URLs follow the pattern:
```
https://www.goodreads.com/author/show/<author-id>
```

Author IDs are not returned by search — visit the book page or use the recorder to find them.

### Post a discussion reply referencing a book

1. Search for the book to get its ID
2. Post with the `--book` flag:
   ```bash
   ./goodreads post-reply 1585066 --message "You should read this!" --book 228233676
   ```
   This inserts a book reference link into the comment automatically.

## Notes

- All shelf and discussion commands launch a headless Chrome instance — they take a few seconds
- Search uses plain HTTP and is fast (no browser needed)
- The session file stores browser cookies in rod format — both the browser commands and the HTTP search client can read it
- If a command fails with "context deadline exceeded", the page layout may have changed — use `--no-headless` to inspect
