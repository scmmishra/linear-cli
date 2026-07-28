package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/chatwoot/linear-cli/internal/cmd"
	kongcompletion "github.com/jotaen/kong-completion"
)

var version = "dev"

// id-first dispatch: users say `linear issue ENG-123 comments` but Kong
// can't have arg:"" siblings to cmd:"". rewriteIDFirstGrammar swaps the id
// and verb tokens so Kong sees its preferred verb-first form.
var (
	issueVerbs   = []string{"view", "comments"}
	commentVerbs = []string{"view"}

	contextNouns = map[string][]string{
		"issue":    issueVerbs,
		"comment":  commentVerbs,
		"doc":      {"view"},
		"document": {"view"},
	}

	helpVerbSwap = regexp.MustCompile(`\b(view|comments)\s+<id>`)
)

var valueTakingGlobalFlags = map[string]bool{
	"--output": true,
	"-o":       true,
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"--help"}
	}
	args = rewriteIDFirstGrammar(args)

	var cli cmd.CLI
	parser := kong.Must(&cli,
		kong.Name("linear"),
		kong.Description("CLI for Linear — fetch issues and comments."),
		kong.Vars{"version": version},
		kong.UsageOnError(),
		kong.Help(idFirstHelpPrinter),
	)

	// Enable shell completions (must be called before Parse)
	kongcompletion.Register(parser)

	ctx, err := parser.Parse(args)
	parser.FatalIfErrorf(err)

	// Commands that don't require a stored credential. `auth status`, `me`,
	// and `whoami` resolve credentials themselves and report "not logged in"
	// gracefully.
	cmdStr := ctx.Command()
	skipAuth := strings.HasPrefix(cmdStr, "auth") ||
		strings.HasPrefix(cmdStr, "completion") ||
		cmdStr == "me" ||
		cmdStr == "whoami" ||
		cmdStr == "version"

	app, err := cmd.NewApp(&cli, skipAuth, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := ctx.Run(app); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// rewriteIDFirstGrammar swaps `<noun> <id> <verb>` to `<noun> <verb> <id>`
// when the args match a known context-noun grammar. Other shapes pass through
// unchanged, so verb-first input still works.
func rewriteIDFirstGrammar(args []string) []string {
	i := firstNonFlagArg(args)
	if i >= len(args) {
		return args
	}
	verbs, ok := contextNouns[args[i]]
	if !ok || i+2 >= len(args) {
		return args
	}
	id := args[i+1]
	if strings.HasPrefix(id, "-") {
		return args
	}
	// A verb in the id slot means the input is already verb-first.
	if slices.Contains(verbs, id) {
		return args
	}
	if !slices.Contains(verbs, args[i+2]) {
		return args
	}
	out := slices.Clone(args)
	out[i+1], out[i+2] = out[i+2], out[i+1]
	return out
}

func firstNonFlagArg(args []string) int {
	i := 0
	for i < len(args) && strings.HasPrefix(args[i], "-") {
		arg := args[i]
		if arg == "--" {
			return i + 1
		}
		if flag, _, hasValue := strings.Cut(arg, "="); hasValue {
			arg = flag
		}
		i++
		if valueTakingGlobalFlags[arg] && !strings.Contains(args[i-1], "=") && i < len(args) {
			i++
		}
	}
	return i
}

// idFirstHelpPrinter runs Kong's default printer into a buffer, then swaps
// `<verb> <id>` → `<id> <verb>` so help reads in the user-facing order.
func idFirstHelpPrinter(o kong.HelpOptions, ctx *kong.Context) error {
	var buf bytes.Buffer
	orig := ctx.Stdout
	ctx.Stdout = &buf
	err := kong.DefaultHelpPrinter(o, ctx)
	ctx.Stdout = orig
	if err != nil {
		return err
	}
	_, werr := fmt.Fprint(orig, helpVerbSwap.ReplaceAllString(buf.String(), `<id> $1`))
	return werr
}
