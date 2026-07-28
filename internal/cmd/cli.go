package cmd

import (
	"github.com/alecthomas/kong"
	kongcompletion "github.com/jotaen/kong-completion"
)

// CLI is the root Kong struct defining the entire command tree.
//
// Grammar:
//   - Plural noun = list:     `linear issues`
//   - Singular + id + verb:   `linear issue ENG-123 comments`
//   - `issue ENG-123` is shorthand for `issue view ENG-123` (default subcommand).
type CLI struct {
	Output  string `short:"o" default:"text" enum:"text,json,csv" help:"Output format."`
	Quiet   bool   `short:"q" help:"Print only IDs."`
	Verbose bool   `short:"v" help:"Show request/response details."`

	// Plurals — list commands.
	Issues IssuesCmd `cmd:"" help:"List issues (yours by default)."`
	Docs   DocsCmd   `cmd:"" aliases:"documents" help:"List documents."`

	// Singulars — context for verbs and subresources.
	Issue   IssueCmd   `cmd:"" help:"View an issue or its comments."`
	Comment CommentCmd `cmd:"" help:"View a comment."`
	Doc     DocCmd     `cmd:"" aliases:"document" help:"View a document."`

	// Workflow. `me` and `whoami` are aliases of `auth status`.
	Me     MeCmd     `cmd:"" help:"Show identity and connection (alias of 'auth status')."`
	Whoami WhoamiCmd `cmd:"" help:"Show identity and connection (alias of 'auth status')."`
	Api    ApiCmd    `cmd:"" help:"Run a raw GraphQL query against the Linear API."`

	// Setup.
	Auth AuthCmd `cmd:"" help:"Login, logout, and status."`

	Completion kongcompletion.Completion `cmd:"" help:"Print shell completion setup."`
	Version    VersionCmd                `cmd:"" help:"Print the CLI version."`

	VersionFlag kong.VersionFlag `name:"version" help:"Print the CLI version."`
}
