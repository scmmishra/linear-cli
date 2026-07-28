package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chatwoot/linear-cli/internal/config"
	"github.com/chatwoot/linear-cli/internal/output"
	"golang.org/x/term"
)

type AuthCmd struct {
	Login  AuthLoginCmd  `cmd:"" help:"Save a Linear personal API key."`
	Logout AuthLogoutCmd `cmd:"" help:"Remove the saved API key."`
	Status AuthStatusCmd `cmd:"" help:"Show identity and connection."`
}

// -----------------------------------------------------------------------------
// auth login
// -----------------------------------------------------------------------------

type AuthLoginCmd struct {
	APIKey string `help:"Personal API key (prompted for if omitted)."`
}

func (c *AuthLoginCmd) Run(app *App) error {
	apiKey := strings.TrimSpace(c.APIKey)
	if apiKey == "" {
		var err error
		apiKey, err = promptAPIKey()
		if err != nil {
			return err
		}
	}
	if apiKey == "" {
		return fmt.Errorf("api key is required")
	}

	// Validate before saving so a typo'd key never lands in the keyring.
	client := newClient(apiKey, false)
	viewer, err := client.Viewer().Get()
	if err != nil {
		return fmt.Errorf("failed to verify API key: %w", err)
	}

	if err := config.SaveAPIKey(apiKey); err != nil {
		return err
	}

	fmt.Printf("Logged in as %s (%s) in workspace %s.\n",
		output.SanitizeText(viewer.DisplayName), output.SanitizeText(viewer.Email),
		output.SanitizeText(viewer.Organization.Name))
	return nil
}

func promptAPIKey() (string, error) {
	fmt.Fprint(os.Stderr, "Create a personal API key at https://linear.app/settings/account/security\n")
	fmt.Fprint(os.Stderr, "API key: ")
	if term.IsTerminal(int(os.Stdin.Fd())) {
		key, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", fmt.Errorf("failed to read API key: %w", err)
		}
		return strings.TrimSpace(string(key)), nil
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("failed to read API key: %w", err)
	}
	return strings.TrimSpace(line), nil
}

// -----------------------------------------------------------------------------
// auth logout
// -----------------------------------------------------------------------------

type AuthLogoutCmd struct{}

func (c *AuthLogoutCmd) Run(app *App) error {
	if err := config.DeleteAPIKey(); err != nil {
		return err
	}
	fmt.Println("Logged out.")
	return nil
}

// -----------------------------------------------------------------------------
// auth status (+ `me` / `whoami` aliases)
// -----------------------------------------------------------------------------

type AuthStatusCmd struct{}

func (c *AuthStatusCmd) Run(app *App) error {
	apiKey, source, err := config.ResolveAPIKey()
	if err != nil {
		fmt.Println("Not logged in. Run 'linear auth login' or set LINEAR_API_KEY.")
		return nil
	}

	client := newClient(apiKey, false)
	viewer, verr := client.Viewer().Get()
	if verr != nil {
		return fmt.Errorf("credential found (%s) but the API rejected it: %w", source, verr)
	}

	if app.Printer.Format == "json" && !app.Printer.Quiet {
		app.Printer.PrintJSON(map[string]any{
			"viewer":            viewer,
			"credential_source": string(source),
		})
		return nil
	}

	app.Printer.PrintDetail([]output.KeyValue{
		{Key: "User", Value: viewer.DisplayName},
		{Key: "Email", Value: viewer.Email},
		{Key: "Workspace", Value: fmt.Sprintf("%s (%s)", viewer.Organization.Name, viewer.Organization.URLKey)},
		{Key: "Credential", Value: string(source)},
	})
	return nil
}

type MeCmd struct{}

func (c *MeCmd) Run(app *App) error { return (&AuthStatusCmd{}).Run(app) }

type WhoamiCmd struct{}

func (c *WhoamiCmd) Run(app *App) error { return (&AuthStatusCmd{}).Run(app) }
