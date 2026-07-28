# linear-cli

A small CLI for [Linear](https://linear.app), focused on fetching issues and comments from the terminal, scripts, and agent workflows.

```bash
linear issue ENG-123                 # view an issue
linear issue ENG-123 --comments      # issue + its comments
linear issue ENG-123 comments        # just the comments
linear comment <uuid>                # one comment by UUID
linear comment '<issue-url>#comment-1a2b3c4d'   # one comment by URL
linear issues                        # your assigned issues
linear issues -s "payment bug"       # search workspace issues by title
```

Issue references can be an identifier (`ENG-123`), a UUID, or a full Linear URL — `linear issue https://linear.app/acme/issue/ENG-123/fix-the-thing` works.

## Install

```bash
go build -o linear ./cmd/linear/
mv linear /usr/local/bin/   # or anywhere on PATH
```

## Authentication

Create a personal API key at <https://linear.app/settings/account/security>, then:

```bash
linear auth login      # prompts for the key (hidden), verifies it, stores it in the OS keyring
linear auth status     # who am I / which workspace (aliases: linear me, linear whoami)
linear auth logout
```

For CI and agents, set `LINEAR_API_KEY` instead — it takes precedence over the keyring.

## Output formats

- `-o text` (default) — tables and key/value details
- `-o json` — full-fidelity API response
- `-o csv` — tables as CSV
- `-q` — quiet mode, print only IDs (pipe-friendly)

## Escape hatch

Anything the CLI doesn't cover:

```bash
linear api 'query { viewer { name } }'
linear api 'query($id: String!) { issue(id: $id) { title } }' --var id=ENG-123
echo 'query { viewer { id } }' | linear api -
```

`--var` values that parse as JSON (numbers, booleans, objects) are passed through typed; everything else is a string.

## Environment variables

| Variable | Purpose |
| --- | --- |
| `LINEAR_API_KEY` | API key override (beats the keyring) |
| `LINEAR_GRAPHQL_ENDPOINT` | GraphQL endpoint override (tests, proxies); defaults to `https://api.linear.app/graphql` |

## Shell completion

```bash
linear completion   # prints setup instructions for your shell
```
