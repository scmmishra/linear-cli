package sdk

type DocumentsService struct {
	client *Client
}

type Document struct {
	ID         string   `json:"id"`
	SlugID     string   `json:"slugId"`
	Title      string   `json:"title"`
	Icon       string   `json:"icon"`
	Content    string   `json:"content"`
	URL        string   `json:"url"`
	Creator    *User    `json:"creator"`
	Project    *Project `json:"project"`
	Initiative *struct {
		Name string `json:"name"`
	} `json:"initiative"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

const documentFields = `
	id
	slugId
	title
	icon
	content
	url
	creator { name displayName }
	project { name }
	initiative { name }
	createdAt
	updatedAt
`

// Get fetches one document by UUID or slug id (the trailing token of a
// document URL).
func (s *DocumentsService) Get(id string) (*Document, error) {
	query := `query($id: String!) { document(id: $id) {` + documentFields + `} }`

	var resp struct {
		Document *Document `json:"document"`
	}
	if err := s.client.Do(query, map[string]any{"id": id}, &resp); err != nil {
		return nil, err
	}
	return resp.Document, nil
}

type DocumentSummary struct {
	ID        string   `json:"id"`
	SlugID    string   `json:"slugId"`
	Title     string   `json:"title"`
	Project   *Project `json:"project"`
	Creator   *User    `json:"creator"`
	UpdatedAt string   `json:"updatedAt"`
}

const documentSummaryFields = `
	id
	slugId
	title
	project { name }
	creator { name displayName }
	updatedAt
`

// List lists workspace documents, most recently updated first. A non-empty
// term filters by title (case insensitive).
func (s *DocumentsService) List(term string, limit int) ([]DocumentSummary, error) {
	query := `query($first: Int!, $filter: DocumentFilter) {
		documents(first: $first, filter: $filter, orderBy: updatedAt) {
			nodes {` + documentSummaryFields + `}
		}
	}`

	variables := map[string]any{"first": limit}
	if term != "" {
		variables["filter"] = map[string]any{
			"title": map[string]any{"containsIgnoreCase": term},
		}
	}

	var resp struct {
		Documents struct {
			Nodes []DocumentSummary `json:"nodes"`
		} `json:"documents"`
	}
	if err := s.client.Do(query, variables, &resp); err != nil {
		return nil, err
	}
	return resp.Documents.Nodes, nil
}
