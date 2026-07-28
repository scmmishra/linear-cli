package sdk

type IssuesService struct {
	client *Client
}

type WorkflowState struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type User struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email,omitempty"`
}

// Label returns the best human-readable name for a user.
func (u *User) Label() string {
	if u == nil {
		return ""
	}
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Name
}

type Team struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type Project struct {
	Name string `json:"name"`
}

type LabelNode struct {
	Name string `json:"name"`
}

type Issue struct {
	ID            string        `json:"id"`
	Identifier    string        `json:"identifier"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	URL           string        `json:"url"`
	BranchName    string        `json:"branchName"`
	PriorityLabel string        `json:"priorityLabel"`
	State         WorkflowState `json:"state"`
	Assignee      *User         `json:"assignee"`
	Creator       *User         `json:"creator"`
	Team          *Team         `json:"team"`
	Project       *Project      `json:"project"`
	Labels        struct {
		Nodes []LabelNode `json:"nodes"`
	} `json:"labels"`
	Comments struct {
		Nodes []Comment `json:"nodes"`
	} `json:"comments"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

const issueFields = `
	id
	identifier
	title
	description
	url
	branchName
	priorityLabel
	state { name type }
	assignee { name displayName }
	creator { name displayName }
	team { key name }
	project { name }
	labels { nodes { name } }
	createdAt
	updatedAt
`

const commentFields = `
	id
	body
	url
	createdAt
	user { name displayName }
	botActor { name }
`

// Get fetches one issue by UUID or identifier (e.g. "ENG-123").
// withComments additionally fetches up to 250 comments, oldest first.
func (s *IssuesService) Get(id string, withComments bool) (*Issue, error) {
	query := `query($id: String!) { issue(id: $id) {` + issueFields + `} }`
	if withComments {
		query = `query($id: String!) { issue(id: $id) {` + issueFields +
			`comments(first: 250) { nodes {` + commentFields + `} } } }`
	}

	var resp struct {
		Issue *Issue `json:"issue"`
	}
	if err := s.client.Do(query, map[string]any{"id": id}, &resp); err != nil {
		return nil, err
	}
	return resp.Issue, nil
}

type IssueSummary struct {
	Identifier    string        `json:"identifier"`
	Title         string        `json:"title"`
	PriorityLabel string        `json:"priorityLabel"`
	State         WorkflowState `json:"state"`
	Assignee      *User         `json:"assignee"`
	UpdatedAt     string        `json:"updatedAt"`
}

const issueSummaryFields = `
	identifier
	title
	priorityLabel
	state { name type }
	assignee { name displayName }
	updatedAt
`

// ListMine lists issues assigned to the authenticated user, most recently
// updated first.
func (s *IssuesService) ListMine(limit int) ([]IssueSummary, error) {
	query := `query($first: Int!) {
		viewer {
			assignedIssues(first: $first, orderBy: updatedAt) {
				nodes {` + issueSummaryFields + `}
			}
		}
	}`

	var resp struct {
		Viewer struct {
			AssignedIssues struct {
				Nodes []IssueSummary `json:"nodes"`
			} `json:"assignedIssues"`
		} `json:"viewer"`
	}
	if err := s.client.Do(query, map[string]any{"first": limit}, &resp); err != nil {
		return nil, err
	}
	return resp.Viewer.AssignedIssues.Nodes, nil
}

// Search lists workspace issues whose title contains the term (case
// insensitive), most recently updated first. An empty term lists all issues.
func (s *IssuesService) Search(term string, limit int) ([]IssueSummary, error) {
	query := `query($first: Int!, $filter: IssueFilter) {
		issues(first: $first, filter: $filter, orderBy: updatedAt) {
			nodes {` + issueSummaryFields + `}
		}
	}`

	variables := map[string]any{"first": limit}
	if term != "" {
		variables["filter"] = map[string]any{
			"title": map[string]any{"containsIgnoreCase": term},
		}
	}

	var resp struct {
		Issues struct {
			Nodes []IssueSummary `json:"nodes"`
		} `json:"issues"`
	}
	if err := s.client.Do(query, variables, &resp); err != nil {
		return nil, err
	}
	return resp.Issues.Nodes, nil
}
