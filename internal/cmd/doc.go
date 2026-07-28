package cmd

import (
	"fmt"

	"github.com/scmmishra/linear-cli/internal/output"
)

// -----------------------------------------------------------------------------
// Plural: `linear docs` — list documents.
// -----------------------------------------------------------------------------

type DocsCmd struct {
	Search string `short:"s" help:"Filter documents by title."`
	Limit  int    `short:"n" default:"25" help:"Maximum number of documents."`
}

func (c *DocsCmd) Run(app *App) error {
	docs, err := app.Client.Documents().List(c.Search, c.Limit)
	if err != nil {
		return err
	}

	if app.Printer.Format == "json" && !app.Printer.Quiet {
		app.Printer.PrintJSON(docs)
		return nil
	}

	if len(docs) == 0 {
		fmt.Println("No documents found.")
		return nil
	}

	headers := []string{"ID", "Title", "Project", "Creator", "Updated"}
	rows := make([][]string, 0, len(docs))
	for _, doc := range docs {
		project := ""
		if doc.Project != nil {
			project = doc.Project.Name
		}
		rows = append(rows, []string{
			doc.SlugID,
			doc.Title,
			project,
			doc.Creator.Label(),
			formatTime(doc.UpdatedAt),
		})
	}

	app.Printer.PrintTable(headers, rows)
	return nil
}

// -----------------------------------------------------------------------------
// Singular: `linear doc <id-or-url>` — view one document.
// -----------------------------------------------------------------------------

type DocCmd struct {
	View DocViewCmd `cmd:"" default:"withargs" help:"View a document (default)."`
}

type DocViewCmd struct {
	ID string `arg:"" help:"Document UUID, slug id, or Linear document URL."`
}

func (c *DocViewCmd) Run(app *App) error {
	id, err := ParseDocRef(c.ID)
	if err != nil {
		return err
	}

	doc, err := app.Client.Documents().Get(id)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("document %s not found", id)
	}

	if app.Printer.Format == "json" && !app.Printer.Quiet {
		app.Printer.PrintJSON(doc)
		return nil
	}
	if app.Printer.Quiet {
		fmt.Fprintln(app.Printer.Writer, doc.SlugID)
		return nil
	}

	pairs := []output.KeyValue{
		{Key: "Title", Value: doc.Title},
		{Key: "Creator", Value: valueOrDash(doc.Creator.Label())},
	}
	if doc.Project != nil {
		pairs = append(pairs, output.KeyValue{Key: "Project", Value: doc.Project.Name})
	}
	if doc.Initiative != nil {
		pairs = append(pairs, output.KeyValue{Key: "Initiative", Value: doc.Initiative.Name})
	}
	pairs = append(pairs,
		output.KeyValue{Key: "Created", Value: formatTime(doc.CreatedAt)},
		output.KeyValue{Key: "Updated", Value: formatTime(doc.UpdatedAt)},
		output.KeyValue{Key: "URL", Value: doc.URL},
	)
	app.Printer.PrintDetail(pairs)

	if doc.Content != "" {
		fmt.Fprintln(app.Printer.Writer)
		app.Printer.PrintBlock(doc.Content)
	}
	return nil
}
