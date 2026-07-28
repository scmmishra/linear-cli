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
}

type IssueViewCmd struct {
	ID       string `arg:"" help:"Issue identifier (ENG-123), UUID, or Linear URL."`
	Comments bool   `short:"c" help:"Also show the issue's comments."`
}

func (c *IssueViewCmd) Run(app *App) error {
	id, err := ParseIssueRef(c.ID)
	if err != nil {
		return err
	}

	issue, err := app.Client.Issues().Get(id, c.Comments)
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

	if c.Comments {
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
