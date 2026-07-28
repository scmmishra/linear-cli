package cmd

import (
	"fmt"

	"github.com/scmmishra/linear-cli/internal/output"
	"github.com/scmmishra/linear-cli/internal/sdk"
)

// -----------------------------------------------------------------------------
// Singular: `linear comment <uuid-or-url>` — view one comment.
// -----------------------------------------------------------------------------

type CommentCmd struct {
	View CommentViewCmd `cmd:"" default:"withargs" help:"View a comment (default)."`
}

type CommentViewCmd struct {
	ID string `arg:"" help:"Comment UUID, or a Linear issue URL with a #comment-… fragment."`
}

func (c *CommentViewCmd) Run(app *App) error {
	ref, err := ParseCommentRef(c.ID)
	if err != nil {
		return err
	}

	var comment *sdk.Comment
	if ref.CommentID != "" {
		comment, err = app.Client.Comments().Get(ref.CommentID)
		if err != nil {
			return err
		}
	} else {
		// URL fragments carry only a short hash of the comment id, which the
		// API can't look up directly — resolve it via the issue's comments.
		issue, comments, lerr := app.Client.Comments().ListForIssue(ref.IssueID)
		if lerr != nil {
			return lerr
		}
		if issue == nil {
			return fmt.Errorf("issue %s not found", ref.IssueID)
		}
		for i := range comments {
			if MatchesHashPrefix(comments[i].ID, ref.HashPrefix) {
				comment = &comments[i]
				comment.Issue = &struct {
					Identifier string `json:"identifier"`
					Title      string `json:"title"`
				}{Identifier: issue.Identifier, Title: issue.Title}
				break
			}
		}
		if comment == nil {
			return fmt.Errorf("no comment matching %q found on %s", ref.HashPrefix, issue.Identifier)
		}
	}

	if comment == nil {
		return fmt.Errorf("comment not found")
	}

	if app.Printer.Format == "json" && !app.Printer.Quiet {
		app.Printer.PrintJSON(comment)
		return nil
	}
	if app.Printer.Quiet {
		fmt.Fprintln(app.Printer.Writer, comment.ID)
		return nil
	}

	pairs := []output.KeyValue{
		{Key: "ID", Value: comment.ID},
		{Key: "Author", Value: comment.Author()},
		{Key: "Created", Value: formatTime(comment.CreatedAt)},
	}
	if comment.Issue != nil {
		pairs = append(pairs, output.KeyValue{
			Key:   "Issue",
			Value: fmt.Sprintf("%s: %s", comment.Issue.Identifier, comment.Issue.Title),
		})
	}
	if comment.URL != "" {
		pairs = append(pairs, output.KeyValue{Key: "URL", Value: comment.URL})
	}
	app.Printer.PrintDetail(pairs)

	fmt.Fprintln(app.Printer.Writer)
	app.Printer.PrintBlock(comment.Body)
	return nil
}
