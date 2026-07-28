package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// ApiCmd runs an arbitrary GraphQL query, for anything the CLI doesn't cover.
//
//	linear api 'query { viewer { name } }'
//	linear api 'query($id: String!) { issue(id: $id) { title } }' --var id=ENG-123
//	echo 'query { viewer { id } }' | linear api -
type ApiCmd struct {
	Query string            `arg:"" help:"GraphQL query, or '-' to read it from stdin."`
	Var   map[string]string `help:"Query variables as key=value (repeatable)." mapsep:","`
}

func (c *ApiCmd) Run(app *App) error {
	query := c.Query
	if query == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read query from stdin: %w", err)
		}
		query = string(data)
	}
	if strings.TrimSpace(query) == "" {
		return fmt.Errorf("query is required")
	}

	variables := make(map[string]any, len(c.Var))
	for k, v := range c.Var {
		// Pass numbers/booleans/objects through as JSON, everything else as string.
		var parsed any
		if err := json.Unmarshal([]byte(v), &parsed); err == nil {
			variables[k] = parsed
		} else {
			variables[k] = v
		}
	}

	raw, err := app.Client.DoRaw(query, variables)
	if err != nil {
		return err
	}

	var pretty any
	if err := json.Unmarshal(raw, &pretty); err != nil {
		fmt.Fprintln(app.Printer.Writer, string(raw))
		return nil
	}
	app.Printer.PrintJSON(pretty)
	return nil
}
