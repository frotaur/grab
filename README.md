# grab

A headless web reader for terminals and AI agents.

## The problem

Models get URLs from docs, error messages, READMEs. They can't open them. If they `curl` raw, they get HTML soup that eats 10k+ tokens and is mostly nav bars and footers.

## What it does

Downloads a URL, extracts the main content, and outputs clean plain text.

```bash
grab https://docs.stripe.com/api/charges
grab https://github.com/user/repo/issues/42
grab https://stackoverflow.com/questions/12345
grab --tokens 2000 https://docs.python.org/3/library/json.html
grab --list https://docs.foo.com/api
```

## Install

```bash
curl -sfL https://raw.githubusercontent.com/frotaur/grab/latest/install.sh | bash
```

Or build from source (requires Go 1.21+):
```bash
git clone https://github.com/frotaur/grab
cd grab
go build -o grab .
mv grab ~/.local/bin/
```

## Features

- **Content extraction** — Strips nav bars, footers, ads, and extracts just the article content.
- **Site-specific extractors** — Optimized for GitHub issues/PRs, Stack Overflow questions, with more to come.
- **Token limiting** — `--tokens 2000` truncates intelligently at paragraph boundaries.
- **Link listing** — `--list` returns just the links on a page, letting you navigate docs without a browser.
- **Local caching** — Results cached in `~/.grab/cache/` with 1-hour TTL. Avoids re-fetching in loops.
- **Plain text output** — No ANSI codes, no TUI. Just clean text that works in pipes, scripts, and AI agent sandboxes.

## Usage

```
grab [flags] URL

Flags:
  -t, --tokens int      truncate output to approximately this many tokens (0 = no limit)
  -l, --list             list links found on the page instead of content
  -r, --raw              output raw extracted text without cleaning
      --no-cache         bypass the local cache
      --user-agent str   HTTP User-Agent header
  -h, --help             help for grab
```

## License

MIT
