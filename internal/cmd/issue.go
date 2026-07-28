package cmd

import (
	"fmt"
	"strings"

	"github.com/chatwoot/linear-cli/internal/output"
	"github.com/chatwoot/linear-cli/internal/sdk"
)

// -----------------------------------------------------------------------------
// Plural: `linear issues` — list issues.
// -----------------------------------------------------------------------------

type IssuesCmd struct {
	Search string `short:"s" help:"Search workspace issues by title instead of listing yours."`
	All    bool   `help:"List all workspace issues instead of yours."`
	Limit  int    `short:"n" default:"25" help:"Maximum number of issues."`
}

func (c *IssuesCmd) Run(app *App) error {
	var (
		issues []sdk.IssueSummary
		err    error
	)
	if c.Search != "" || c.All {
		issues, err = app.Client.Issues().Search(c.Search, c.Limit)
	} else {
		issues, err = app.Client.Issues().ListMine(c.Limit)
	}
	if err != nil {
		return err
	}

	if app.Printer.Format == "json" && !app.Printer.Quiet {
		app.Printer.PrintJSON(issues)
		return nil
	}

	if len(issues) == 0 {
		fmt.Println("No issues found.")
		return nil
	}

	headers := []string{"ID", "Title", "State", "Priority", "Assignee", "Updated"}
	rows := make([][]string, 0, len(issues))
	for _, issue := range issues {
		rows = append(rows, []string{
			issue.Identifier,
			issue.Title,
			issue.State.Name,
			issue.PriorityLabel,
			issue.Assignee.Label(),
			formatTime(issue.UpdatedAt),
		})
	}

	app.Printer.PrintTable(headers, rows)
	return nil
}

// -----------------------------------------------------------------------------
// Singular: `linear issue ENG-123 [view|comments]` — one issue.
// -----------------------------------------------------------------------------

type IssueCmd struct {
	View     IssueViewCmd     `cmd:"" default:"withargs" help:"View an issue (default)."`
	Comments IssueCommentsCmd `cmd:"" help:"List an issue's comments."`
	Docs     IssueDocsCmd     `cmd:"" help:"List the documents of an issue's project."`
}

type IssueViewCmd struct {
	ID       string `arg:"" help:"Issue identifier (ENG-123), UUID, or Linear URL."`
	Comments bool   `short:"c" help:"Also show the issue's comments."`
	All      bool   `help:"Also show the issue's comments and its project's documents."`
}

func (c *IssueViewCmd) Run(app *App) error {
	id, err := ParseIssueRef(c.ID)
	if err != nil {
		return err
	}

	withComments := c.Comments || c.All
	issue, err := app.Client.Issues().Get(id, withComments, c.All)
	if err != nil {
		return err
	}
	if issue == nil {
		return fmt.Errorf("issue %s not found", id)
	}

	if app.Printer.Format == "json" && !app.Printer.Quiet {
		app.Printer.PrintJSON(issue)
		return nil
	}
	if app.Printer.Quiet {
		fmt.Fprintln(app.Printer.Writer, issue.Identifier)
		return nil
	}

	labels := make([]string, 0, len(issue.Labels.Nodes))
	for _, l := range issue.Labels.Nodes {
		labels = append(labels, l.Name)
	}

	pairs := []output.KeyValue{
		{Key: "ID", Value: issue.Identifier},
		{Key: "Title", Value: issue.Title},
		{Key: "State", Value: issue.State.Name},
		{Key: "Priority", Value: issue.PriorityLabel},
		{Key: "Assignee", Value: valueOrDash(issue.Assignee.Label())},
		{Key: "Creator", Value: valueOrDash(issue.Creator.Label())},
	}
	if issue.Team != nil {
		pairs = append(pairs, output.KeyValue{Key: "Team", Value: issue.Team.Name})
	}
	if issue.Project != nil {
		pairs = append(pairs, output.KeyValue{Key: "Project", Value: issue.Project.Name})
	}
	if len(labels) > 0 {
		pairs = append(pairs, output.KeyValue{Key: "Labels", Value: strings.Join(labels, ", ")})
	}
	pairs = append(pairs,
		output.KeyValue{Key: "Created", Value: formatTime(issue.CreatedAt)},
		output.KeyValue{Key: "Updated", Value: formatTime(issue.UpdatedAt)},
		output.KeyValue{Key: "URL", Value: issue.URL},
	)
	app.Printer.PrintDetail(pairs)

	if issue.Description != "" {
		fmt.Fprintln(app.Printer.Writer)
		app.Printer.PrintBlock(issue.Description)
	}

	if c.All && issue.Project != nil && len(issue.Project.Documents.Nodes) > 0 {
		fmt.Fprintf(app.Printer.Writer, "\nDocuments (%s):\n", output.SanitizeText(issue.Project.Name))
		for _, doc := range issue.Project.Documents.Nodes {
			fmt.Fprintf(app.Printer.Writer, "  %s  %s  (updated %s)\n",
				doc.SlugID, output.SanitizeText(doc.Title), formatTime(doc.UpdatedAt))
		}
	}

	if withComments {
		printComments(app, issue.Comments.Nodes)
	}
	return nil
}

type IssueCommentsCmd struct {
	ID string `arg:"" help:"Issue identifier (ENG-123), UUID, or Linear URL."`
}

func (c *IssueCommentsCmd) Run(app *App) error {
	id, err := ParseIssueRef(c.ID)
	if err != nil {
		return err
	}

	issue, comments, err := app.Client.Comments().ListForIssue(id)
	if err != nil {
		return err
	}
	if issue == nil {
		return fmt.Errorf("issue %s not found", id)
	}

	if app.Printer.Format == "json" && !app.Printer.Quiet {
		app.Printer.PrintJSON(comments)
		return nil
	}
	if app.Printer.Quiet {
		for _, comment := range comments {
			fmt.Fprintln(app.Printer.Writer, comment.ID)
		}
		return nil
	}

	if len(comments) == 0 {
		fmt.Printf("No comments on %s.\n", issue.Identifier)
		return nil
	}

	fmt.Fprintf(app.Printer.Writer, "%s: %s (%d comments)\n", issue.Identifier, output.SanitizeText(issue.Title), len(comments))
	printComments(app, comments)
	return nil
}

type IssueDocsCmd struct {
	ID string `arg:"" help:"Issue identifier (ENG-123), UUID, or Linear URL."`
}

func (c *IssueDocsCmd) Run(app *App) error {
	id, err := ParseIssueRef(c.ID)
	if err != nil {
		return err
	}

	issue, err := app.Client.Issues().Get(id, false, true)
	if err != nil {
		return err
	}
	if issue == nil {
		return fmt.Errorf("issue %s not found", id)
	}
	if issue.Project == nil {
		fmt.Printf("%s has no project, so no documents.\n", issue.Identifier)
		return nil
	}

	docs := issue.Project.Documents.Nodes
	if app.Printer.Format == "json" && !app.Printer.Quiet {
		app.Printer.PrintJSON(docs)
		return nil
	}
	if len(docs) == 0 {
		fmt.Printf("No documents in project %s.\n", issue.Project.Name)
		return nil
	}

	headers := []string{"ID", "Title", "Creator", "Updated"}
	rows := make([][]string, 0, len(docs))
	for _, doc := range docs {
		rows = append(rows, []string{
			doc.SlugID,
			doc.Title,
			doc.Creator.Label(),
			formatTime(doc.UpdatedAt),
		})
	}
	app.Printer.PrintTable(headers, rows)
	return nil
}

func printComments(app *App, comments []sdk.Comment) {
	for _, comment := range comments {
		fmt.Fprintln(app.Printer.Writer)
		fmt.Fprintf(app.Printer.Writer, "--- %s · %s · %s\n",
			output.SanitizeText(comment.Author()), formatTime(comment.CreatedAt), comment.ID)
		app.Printer.PrintBlock(comment.Body)
	}
}

func valueOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
