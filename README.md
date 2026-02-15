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

### Post to a discussion

```
./goodreads post 1585066 --message "Hello from the CLI!"
```

Posts a comment to a Goodreads discussion topic. The argument is the topic ID from the URL (e.g. `goodreads.com/topic/show/1585066`).

## How it works

- **Search** uses Goodreads' JSON autocomplete endpoint (`/book/auto_complete?format=json`) via plain HTTP
- **Login** and **shelf operations** use [rod](https://github.com/go-rod/rod) for headless browser automation, since Goodreads routes login through Amazon's OpenID and shelf mutations go through Next.js/React internals
- Session cookies are persisted to `~/.goodreads-cli-session` so you only need to log in once
