package sdk

type CommentsService struct {
	client *Client
}

type BotActor struct {
	Name string `json:"name"`
}

type Comment struct {
	ID        string    `json:"id"`
	Body      string    `json:"body"`
	URL       string    `json:"url"`
	CreatedAt string    `json:"createdAt"`
	User      *User     `json:"user"`
	BotActor  *BotActor `json:"botActor"`
	Issue     *struct {
		Identifier string `json:"identifier"`
		Title      string `json:"title"`
	} `json:"issue,omitempty"`
}

// Author returns the comment's author name, falling back to the bot actor.
func (c *Comment) Author() string {
	if name := c.User.Label(); name != "" {
		return name
	}
	if c.BotActor != nil && c.BotActor.Name != "" {
		return c.BotActor.Name
	}
	return "(unknown)"
}

// Get fetches one comment by UUID.
func (s *CommentsService) Get(id string) (*Comment, error) {
	query := `query($id: String!) {
		comment(id: $id) {` + commentFields + `
			issue { identifier title }
		}
	}`

	var resp struct {
		Comment *Comment `json:"comment"`
	}
	if err := s.client.Do(query, map[string]any{"id": id}, &resp); err != nil {
		return nil, err
	}
	return resp.Comment, nil
}

// ListForIssue fetches an issue's comments, oldest first.
func (s *CommentsService) ListForIssue(issueID string) (*Issue, []Comment, error) {
	issue, err := s.client.Issues().Get(issueID, true, false)
	if err != nil {
		return nil, nil, err
	}
	if issue == nil {
		return nil, nil, nil
	}
	return issue, issue.Comments.Nodes, nil
}
