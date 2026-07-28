package sdk

type ViewerService struct {
	client *Client
}

type Viewer struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	Email        string `json:"email"`
	Organization struct {
		Name   string `json:"name"`
		URLKey string `json:"urlKey"`
	} `json:"organization"`
}

// Get fetches the authenticated user and workspace.
func (s *ViewerService) Get() (*Viewer, error) {
	query := `query {
		viewer {
			id
			name
			displayName
			email
			organization { name urlKey }
		}
	}`

	var resp struct {
		Viewer *Viewer `json:"viewer"`
	}
	if err := s.client.Do(query, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Viewer, nil
}
