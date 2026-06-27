package util

import "strings"

// NormalizeCommentText rewrites a single-line block comment (/* @x ... */) into
// the line-comment form (// @x ...) so annotation regexes that require a "//"
// prefix match it. Line comments and any other text are returned unchanged.
func NormalizeCommentText(text string) string {
	if strings.HasPrefix(text, "/*") {
		inner := strings.TrimSuffix(strings.TrimPrefix(text, "/*"), "*/")
		return "// " + strings.TrimSpace(inner)
	}
	return text
}
