package cmd

import "testing"

func TestParseIssueRef(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "ENG-123", want: "ENG-123"},
		{in: "eng-123", want: "ENG-123"},
		{in: " ENG-123 ", want: "ENG-123"},
		{in: "a3f1c2d4-0000-4000-8000-000000000000", want: "a3f1c2d4-0000-4000-8000-000000000000"},
		{in: "https://linear.app/acme/issue/ENG-123/fix-the-thing", want: "ENG-123"},
		{in: "https://linear.app/acme/issue/ENG-123/fix-the-thing#comment-a3f1c2d4", want: "ENG-123"},
		{in: "not a ref", wantErr: true},
		{in: "https://linear.app/acme/project/foo", wantErr: true},
		{in: "", wantErr: true},
	}

	for _, tt := range tests {
		got, err := ParseIssueRef(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseIssueRef(%q) expected error, got %q", tt.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseIssueRef(%q) unexpected error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseIssueRef(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParseCommentRef(t *testing.T) {
	uuid := "a3f1c2d4-0000-4000-8000-000000000000"

	ref, err := ParseCommentRef(uuid)
	if err != nil {
		t.Fatalf("uuid ref: %v", err)
	}
	if ref.CommentID != uuid {
		t.Errorf("CommentID = %q, want %q", ref.CommentID, uuid)
	}

	ref, err = ParseCommentRef("https://linear.app/acme/issue/ENG-123/slug#comment-a3f1c2d4")
	if err != nil {
		t.Fatalf("url ref: %v", err)
	}
	if ref.IssueID != "ENG-123" || ref.HashPrefix != "a3f1c2d4" {
		t.Errorf("got issue %q hash %q", ref.IssueID, ref.HashPrefix)
	}

	if _, err := ParseCommentRef("https://linear.app/acme/issue/ENG-123/slug"); err == nil {
		t.Error("issue URL without fragment should be rejected")
	}
	if _, err := ParseCommentRef("ENG-123"); err == nil {
		t.Error("bare identifier should be rejected as a comment ref")
	}
}

func TestParseDocRef(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "a3f1c2d4-0000-4000-8000-000000000000", want: "a3f1c2d4-0000-4000-8000-000000000000"},
		{in: "3f1c2d4a5b6e", want: "3f1c2d4a5b6e"},
		{in: "https://linear.app/acme/document/roadmap-notes-3f1c2d4a5b6e", want: "3f1c2d4a5b6e"},
		{in: "https://linear.app/acme/document/3f1c2d4a5b6e", want: "3f1c2d4a5b6e"},
		{in: "https://linear.app/acme/issue/ENG-123/slug", wantErr: true},
		{in: "", wantErr: true},
	}

	for _, tt := range tests {
		got, err := ParseDocRef(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseDocRef(%q) expected error, got %q", tt.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseDocRef(%q) unexpected error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseDocRef(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestMatchesHashPrefix(t *testing.T) {
	uuid := "a3f1c2d4-0000-4000-8000-000000000000"
	if !MatchesHashPrefix(uuid, "a3f1c2d4") {
		t.Error("expected prefix match")
	}
	if MatchesHashPrefix(uuid, "deadbeef") {
		t.Error("unexpected prefix match")
	}
}
