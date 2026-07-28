# linear-cli

A small, read-focused CLI for [Linear](https://linear.app) — fetch issues, comments, and documents from the terminal, scripts, and agent workflows.

```bash
linear issue ENG-123                 # view an issue
linear issue ENG-123 --comments      # issue + its comments
linear issue ENG-123 --all           # issue + comments + its project's documents
linear issue ENG-123 comments        # just the comments
linear issue ENG-123 docs            # just the project's documents
linear comment <uuid-or-url>         # view one comment
linear doc <id-or-url>               # view a document (prints its markdown)
linear issues                        # your assigned issues (-s to search, --all for workspace)
linear docs                          # list documents
```

References are flexible: issues take an identifier (`ENG-123`), UUID, or pasted Linear URL; comments take a UUID or an issue URL with a `#comment-…` fragment; documents take a UUID, slug id, or document URL.

## Install

Requires Go 1.25+ (builds from source):

```bash
curl -fsSL https://raw.githubusercontent.com/scmmishra/linear-cli/main/install.sh | sh
```

Installs to `~/.local/bin` (override with `LINEAR_INSTALL_DIR`, pin with `LINEAR_VERSION`). From a checkout, `./install.sh` builds the local source, or just:

```bash
go install github.com/scmmishra/linear-cli/cmd/linear@latest
```

## Authentication

Create a personal API key at [linear.app/settings/account/security](https://linear.app/settings/account/security), then:

```bash
linear auth login      # prompts for the key, verifies it, stores it in the OS keyring
linear me              # who am I / which workspace
```

For CI and agents, set `LINEAR_API_KEY` instead — it takes precedence over the keyring.

## Output

`-o json` for full-fidelity JSON, `-o csv` for tables as CSV, `-q` for IDs only (pipe-friendly). Default text output is for humans.

Anything the commands don't cover:

```bash
linear api 'query($id: String!) { issue(id: $id) { title } }' --var id=ENG-123
```

## Agent skill

The repo ships a [skill](skills/linear-cli/SKILL.md) that teaches coding agents the CLI's grammar, output contract, and trust rules. Install it with [skills.sh](https://skills.sh):

```bash
npx skills add https://github.com/scmmishra/linear-cli --skill linear-cli
```

## License

[MIT](LICENSE)
