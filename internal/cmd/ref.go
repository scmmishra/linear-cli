package cmd

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	identifierRe  = regexp.MustCompile(`^[A-Za-z0-9]+-\d+$`)
	uuidRe        = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	commentHashRe = regexp.MustCompile(`^comment-([0-9a-fA-F]+)$`)
)

// ParseIssueRef normalizes an issue reference — an identifier ("ENG-123"), a
// UUID, or a Linear URL (https://linear.app/<ws>/issue/ENG-123/slug) — into
// the id string the API accepts.
func ParseIssueRef(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if uuidRe.MatchString(ref) {
		return ref, nil
	}
	if identifierRe.MatchString(ref) {
		return strings.ToUpper(ref), nil
	}
	if id, _, err := parseIssueURL(ref); err == nil {
		return id, nil
	}
	return "", fmt.Errorf("invalid issue reference %q: expected an identifier like ENG-123, a UUID, or a Linear issue URL", ref)
}

// CommentRef is a parsed comment reference: either a direct comment UUID, or
// an issue plus a short comment-id prefix taken from a URL fragment
// (…/issue/ENG-123/slug#comment-1a2b3c4d).
type CommentRef struct {
	CommentID  string // set when the ref is a UUID
	IssueID    string // set when the ref is an issue URL with a #comment- fragment
	HashPrefix string // hex prefix of the comment UUID, from the fragment
}

// ParseCommentRef accepts a comment UUID or a Linear issue URL with a
// #comment-<hash> fragment.
func ParseCommentRef(ref string) (*CommentRef, error) {
	ref = strings.TrimSpace(ref)
	if uuidRe.MatchString(ref) {
		return &CommentRef{CommentID: ref}, nil
	}
	if id, fragment, err := parseIssueURL(ref); err == nil {
		if m := commentHashRe.FindStringSubmatch(fragment); m != nil {
			return &CommentRef{IssueID: id, HashPrefix: strings.ToLower(m[1])}, nil
		}
		return nil, fmt.Errorf("URL %q has no #comment-… fragment; pass a comment URL or a comment UUID", ref)
	}
	return nil, fmt.Errorf("invalid comment reference %q: expected a comment UUID or a Linear comment URL", ref)
}

// MatchesHashPrefix reports whether a comment UUID matches a URL-fragment hash
// prefix (compared with dashes removed).
func MatchesHashPrefix(commentUUID, hashPrefix string) bool {
	compact := strings.ToLower(strings.ReplaceAll(commentUUID, "-", ""))
	return strings.HasPrefix(compact, hashPrefix)
}

// parseIssueURL extracts the issue identifier and URL fragment from a Linear
// issue URL: https://linear.app/<workspace>/issue/<IDENTIFIER>/<slug>#fragment
func parseIssueURL(ref string) (id, fragment string, err error) {
	if !strings.Contains(ref, "://") {
		return "", "", fmt.Errorf("not a URL")
	}
	u, err := url.Parse(ref)
	if err != nil {
		return "", "", err
	}
	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, seg := range segments {
		if seg == "issue" && i+1 < len(segments) {
			candidate := segments[i+1]
			if identifierRe.MatchString(candidate) {
				return strings.ToUpper(candidate), u.Fragment, nil
			}
		}
	}
	return "", "", fmt.Errorf("no issue identifier in URL")
}
