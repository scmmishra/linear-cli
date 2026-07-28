package main

import (
	"slices"
	"testing"
)

func TestRewriteIDFirstGrammar(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "issue id-first comments",
			in:   []string{"issue", "ENG-123", "comments"},
			want: []string{"issue", "comments", "ENG-123"},
		},
		{
			name: "issue id-first view",
			in:   []string{"issue", "ENG-123", "view"},
			want: []string{"issue", "view", "ENG-123"},
		},
		{
			name: "issue url id-first",
			in:   []string{"issue", "https://linear.app/acme/issue/ENG-123/slug", "comments"},
			want: []string{"issue", "comments", "https://linear.app/acme/issue/ENG-123/slug"},
		},
		{
			name: "already verb-first untouched",
			in:   []string{"issue", "comments", "ENG-123"},
			want: []string{"issue", "comments", "ENG-123"},
		},
		{
			name: "default view shorthand untouched",
			in:   []string{"issue", "ENG-123"},
			want: []string{"issue", "ENG-123"},
		},
		{
			name: "id followed by flag untouched",
			in:   []string{"issue", "ENG-123", "--comments"},
			want: []string{"issue", "ENG-123", "--comments"},
		},
		{
			name: "unknown noun untouched",
			in:   []string{"issues", "ENG-123", "comments"},
			want: []string{"issues", "ENG-123", "comments"},
		},
		{
			name: "leading global flag with value",
			in:   []string{"-o", "json", "issue", "ENG-123", "comments"},
			want: []string{"-o", "json", "issue", "comments", "ENG-123"},
		},
		{
			name: "leading global flag with equals",
			in:   []string{"--output=json", "issue", "ENG-123", "comments"},
			want: []string{"--output=json", "issue", "comments", "ENG-123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rewriteIDFirstGrammar(tt.in)
			if !slices.Equal(got, tt.want) {
				t.Errorf("rewriteIDFirstGrammar(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
