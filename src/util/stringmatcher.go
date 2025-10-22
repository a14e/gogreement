package util

import ac "github.com/BobuSumisu/aho-corasick"

// we use this for MIT licence

// @immutable
// @constructor NewStringMatcher
type StringMatcher struct {
	trie *ac.Trie
}

func NewStringMatcher(dict []string) *StringMatcher {
	tb := ac.NewTrieBuilder().AddStrings(dict)
	return &StringMatcher{trie: tb.Build()}
}

func (m *StringMatcher) Contains(b []byte) bool {
	return m.trie.MatchFirst(b) != nil
}
