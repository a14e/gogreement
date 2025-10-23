package util

import "github.com/cloudflare/ahocorasick"

// @immutable
// @constructor NewStringMatcher
type StringMatcher struct {
	matcher *ahocorasick.Matcher
}

func NewStringMatcher(dict []string) *StringMatcher {
	m := ahocorasick.NewStringMatcher(dict)
	return &StringMatcher{matcher: m}
}

func (m *StringMatcher) Contains(b []byte) bool {
	return m.matcher.Contains(b)
}
