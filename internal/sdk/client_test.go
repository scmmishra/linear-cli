package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewClient("lin_api_test", WithEndpoint(server.URL), WithHTTPClient(server.Client()))
}

func TestDoSendsAuthAndQuery(t *testing.T) {
	var gotAuth, gotQuery string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		var req graphQLRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		gotQuery = req.Query
		_, _ = w.Write([]byte(`{"data": {"viewer": {"name": "Ada"}}}`))
	})

	var resp struct {
		Viewer struct {
			Name string `json:"name"`
		} `json:"viewer"`
	}
	if err := client.Do(`query { viewer { name } }`, nil, &resp); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "lin_api_test" {
		t.Errorf("Authorization = %q, want raw key", gotAuth)
	}
	if !strings.Contains(gotQuery, "viewer") {
		t.Errorf("query not sent: %q", gotQuery)
	}
	if resp.Viewer.Name != "Ada" {
		t.Errorf("Name = %q", resp.Viewer.Name)
	}
}

func TestDoSurfacesGraphQLErrors(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errors": [{"message": "Entity not found"}]}`))
	})

	err := client.Do(`query { issue(id: "NOPE-1") { id } }`, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "Entity not found") {
		t.Errorf("expected GraphQL error, got %v", err)
	}
}

func TestDoSurfacesHTTPErrors(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`not json`))
	})

	err := client.Do(`query { viewer { id } }`, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 error, got %v", err)
	}
}

func TestIssueGetParsesResponse(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data": {"issue": {
			"identifier": "ENG-123",
			"title": "Fix the thing",
			"state": {"name": "In Progress", "type": "started"},
			"assignee": {"name": "ada", "displayName": "Ada"},
			"comments": {"nodes": [
				{"id": "c1", "body": "hello", "user": {"displayName": "Bob"}},
				{"id": "c2", "body": "beep", "user": null, "botActor": {"name": "CI Bot"}}
			]}
		}}}`))
	})

	issue, err := client.Issues().Get("ENG-123", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if issue.Identifier != "ENG-123" || issue.State.Name != "In Progress" {
		t.Errorf("unexpected issue: %+v", issue)
	}
	if issue.Assignee.Label() != "Ada" {
		t.Errorf("assignee label = %q", issue.Assignee.Label())
	}
	if len(issue.Comments.Nodes) != 2 {
		t.Fatalf("comments = %d", len(issue.Comments.Nodes))
	}
	if got := issue.Comments.Nodes[1].Author(); got != "CI Bot" {
		t.Errorf("bot author = %q", got)
	}
}

func TestEndpointPinRefusesNonLinearHosts(t *testing.T) {
	for _, endpoint := range []string{
		"https://evil.example/graphql",
		"https://api.linear.app.evil.example/graphql",
		"http://169.254.169.254/graphql",
	} {
		client := NewClient("lin_api_test", WithEndpoint(endpoint))
		err := client.Do(`query { viewer { id } }`, nil, nil)
		if err == nil || !strings.Contains(err.Error(), "refusing endpoint") {
			t.Errorf("endpoint %q: expected refusal, got %v", endpoint, err)
		}
	}
}

func TestEndpointPinAllowsLoopbackWithPinnedTransport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data": {"viewer": {"id": "u1"}}}`))
	}))
	t.Cleanup(server.Close)

	// A proxy env var must not reroute traffic: the pinned transport ignores it.
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:1")
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:1")

	// No WithHTTPClient — exercises the real pinned transport.
	client := NewClient("lin_api_test", WithEndpoint(server.URL))
	var resp struct {
		Viewer struct {
			ID string `json:"id"`
		} `json:"viewer"`
	}
	if err := client.Do(`query { viewer { id } }`, nil, &resp); err != nil {
		t.Fatalf("loopback endpoint should work: %v", err)
	}
	if resp.Viewer.ID != "u1" {
		t.Errorf("ID = %q", resp.Viewer.ID)
	}
}

func TestPinnedTransportBlocksOtherHosts(t *testing.T) {
	transport := pinnedTransport(DefaultAPIHost)
	_, err := transport.DialContext(t.Context(), "tcp", "evil.example:443")
	if err == nil || !strings.Contains(err.Error(), "blocked") {
		t.Errorf("expected blocked dial, got %v", err)
	}
}

func TestIssueGetNotFound(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data": {"issue": null}}`))
	})

	issue, err := client.Issues().Get("NOPE-1", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if issue != nil {
		t.Errorf("expected nil issue, got %+v", issue)
	}
}
