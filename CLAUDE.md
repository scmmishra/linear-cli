# linear-cli

CLI for the Linear GraphQL API, focused on fetching issues and comments. Modeled on the sibling `chatwoot-cli` repo — same Kong grammar, printer, and keyring patterns.

## Build & Run

```bash
go build -o linear ./cmd/linear/   # build binary
go test ./...                      # run tests
./linear                           # no args → prints help
./linear issues                    # plural noun = list
./linear issue ENG-123             # `issue ENG-123` = view issue
./linear issue ENG-123 comments    # id-first verb dispatch
```

## Project Structure

```
cmd/linear/main.go         Entry point: pre-parses id-first grammar, then Kong
internal/
  sdk/                     GraphQL client + services (issues, comments, documents, viewer)
  cmd/                     Kong command structs with Run(app *App) error
  config/                  API key in OS keyring (LINEAR_API_KEY env wins)
  output/                  Printer: text (tabwriter), JSON, CSV formats + quiet mode
```

## Architecture

- **CLI framework**: Kong (alecthomas/kong) — struct-based command tree with tags
- **SDK pattern**: `client.Issues()`, `client.Comments()`, `client.Viewer()` return service objects; all go through `client.Do(query, variables, &result)` against a single GraphQL endpoint
- **Command pattern**: each command is a struct with `Run(app *App) error`
- **App struct** holds `Client`, `Printer` — passed to all commands

## Key Conventions

- Grammar: plural noun = list (`IssuesCmd`); singular noun = parent struct with verb subcommands; `default:"withargs"` on View routes `linear issue ENG-123` to `issue view ENG-123`
- `cmd/linear/main.go` runs `rewriteIDFirstGrammar` to swap `<noun> <id> <verb>` → `<noun> <verb> <id>` before Kong parses; IDs are free-form strings (identifiers, UUIDs, URLs), so the rewrite only requires the verb slot to match
- Issue refs are normalized by `ParseIssueRef` (internal/cmd/ref.go): identifier, UUID, or Linear URL all work
- Comment URLs only carry a short `#comment-<hash>` prefix of the comment UUID; `CommentViewCmd` resolves it by fetching the issue's comments and prefix-matching
- No config file: the only state is the API key in the keyring (service `linear-cli`, entry `api-key`); `LINEAR_API_KEY` env always wins
- `skipAuth` in main.go: auth/completion/version commands bypass client creation
- `LINEAR_GRAPHQL_ENDPOINT` env overrides the API endpoint (tests use a local mock server)

## Linear API Notes

- Single endpoint: `POST https://api.linear.app/graphql`
- Personal API keys go in the `Authorization` header **as-is** (no `Bearer` prefix)
- `issue(id:)` accepts both UUIDs and identifiers like `ENG-123`
- `comment(id:)` accepts only UUIDs
- `document(id:)` accepts a UUID or slug id; document URLs end in `<title-slug>-<slugId>` and `ParseDocRef` extracts the trailing slug id token
- Comments may have `user: null` with a `botActor` instead (integrations/bots)
- Errors come back as `{errors: [{message}]}` with HTTP 200 or 4xx

## Commits

Use conventional commits without scope: `feat:`, `fix:`, `chore:`, `refactor:`, `docs:`
