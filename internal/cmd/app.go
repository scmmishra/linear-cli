package cmd

import (
	"fmt"
	"os"

	"github.com/chatwoot/linear-cli/internal/config"
	"github.com/chatwoot/linear-cli/internal/output"
	"github.com/chatwoot/linear-cli/internal/sdk"
)

// App holds shared state passed to every command's Run method.
type App struct {
	Client  *sdk.Client
	Printer *output.Printer
	Version string
}

// NewApp creates an App from the parsed CLI flags. Commands that don't need
// auth (auth login/logout/status, version, completion) pass skipAuth=true.
func NewApp(cli *CLI, skipAuth bool, version string) (*App, error) {
	printer := output.NewPrinter(cli.Output, cli.Quiet)

	if skipAuth {
		return &App{Printer: printer, Version: version}, nil
	}

	apiKey, _, err := config.ResolveAPIKey()
	if err != nil {
		return nil, fmt.Errorf("not authenticated: %w", err)
	}

	client := newClient(apiKey, cli.Verbose)

	return &App{
		Client:  client,
		Printer: printer,
		Version: version,
	}, nil
}

// newClient builds an SDK client, honoring the LINEAR_GRAPHQL_ENDPOINT
// override (tests, proxies). Auth commands that construct their own client
// use this too.
func newClient(apiKey string, verbose bool) *sdk.Client {
	opts := []sdk.ClientOption{sdk.WithVerbose(verbose)}
	if endpoint := os.Getenv("LINEAR_GRAPHQL_ENDPOINT"); endpoint != "" {
		opts = append(opts, sdk.WithEndpoint(endpoint))
	}
	return sdk.NewClient(apiKey, opts...)
}
