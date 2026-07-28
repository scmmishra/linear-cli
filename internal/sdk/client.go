package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/scmmishra/linear-cli/internal/output"
)

const DefaultEndpoint = "https://api.linear.app/graphql"

// Client is a minimal Linear GraphQL client.
type Client struct {
	Endpoint   string
	APIKey     string
	Verbose    bool
	httpClient *http.Client
}

type ClientOption func(*Client)

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func WithEndpoint(endpoint string) ClientOption {
	return func(c *Client) {
		c.Endpoint = endpoint
	}
}

func WithVerbose(verbose bool) ClientOption {
	return func(c *Client) {
		c.Verbose = verbose
	}
}

func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		Endpoint: DefaultEndpoint,
		APIKey:   apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type graphQLError struct {
	Message    string `json:"message"`
	Extensions struct {
		Code            string `json:"code"`
		UserPresentable bool   `json:"userPresentableMessage"`
	} `json:"extensions"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors"`
}

// Do executes a GraphQL query and decodes the `data` object into v.
func (c *Client) Do(query string, variables map[string]any, v any) error {
	payload, err := json.Marshal(graphQLRequest{Query: query, Variables: variables})
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	// Personal API keys go in the Authorization header as-is (no Bearer prefix);
	// OAuth tokens arrive already prefixed by the caller.
	req.Header.Set("Authorization", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	if c.Verbose {
		fmt.Fprintf(os.Stderr, "> POST %s\n> %s\n", c.Endpoint, compactQuery(query))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if c.Verbose {
		fmt.Fprintf(os.Stderr, "< %s\n", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var gqlResp graphQLResponse
	if jerr := json.Unmarshal(body, &gqlResp); jerr != nil {
		if resp.StatusCode >= 400 {
			return fmt.Errorf("API error %d: %s", resp.StatusCode, output.SanitizeText(string(body)))
		}
		return fmt.Errorf("failed to decode response: %w", jerr)
	}

	if len(gqlResp.Errors) > 0 {
		msgs := make([]string, 0, len(gqlResp.Errors))
		for _, e := range gqlResp.Errors {
			msgs = append(msgs, e.Message)
		}
		return fmt.Errorf("API error: %s", output.SanitizeText(strings.Join(msgs, "; ")))
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, output.SanitizeText(string(body)))
	}

	if v == nil {
		return nil
	}
	if err := json.Unmarshal(gqlResp.Data, v); err != nil {
		return fmt.Errorf("failed to decode response data: %w", err)
	}
	return nil
}

// DoRaw executes a query and returns the raw `data` payload, for `linear api`.
func (c *Client) DoRaw(query string, variables map[string]any) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.Do(query, variables, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func compactQuery(q string) string {
	return strings.Join(strings.Fields(q), " ")
}

// Issues returns the issues service.
func (c *Client) Issues() *IssuesService {
	return &IssuesService{client: c}
}

// Documents returns the documents service.
func (c *Client) Documents() *DocumentsService {
	return &DocumentsService{client: c}
}

// Comments returns the comments service.
func (c *Client) Comments() *CommentsService {
	return &CommentsService{client: c}
}

// Viewer returns the viewer (current user) service.
func (c *Client) Viewer() *ViewerService {
	return &ViewerService{client: c}
}
