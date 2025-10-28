package util

import (
	"go/token"
	"strings"

	"gogreement/src/codes"
)

// IgnoreAnnotation interface for adding annotations to the set
// This allows adding from ignore.IgnoreAnnotation without circular dependency
type IgnoreAnnotation interface {
	GetCodes() []string
	GetStartPos() token.Pos
	GetEndPos() token.Pos
}

// IgnoreMarker is the internal storage type for ignore annotations
// Using a concrete struct instead of interface ensures proper gob serialization
// @immutable
type IgnoreMarker struct {
	Codes    []string
	StartPos token.Pos
	EndPos   token.Pos
}

// IgnoreSet provides fast lookup for ignore annotations by code and position
// Can be initialized as &IgnoreSet{}
type IgnoreSet struct {
	// All markers in order they were added
	Markers []IgnoreMarker

	// Index: code -> list of marker indices that contain this code
	CodeIndex map[string][]int

	// Min and max positions for quick range check
	MinPos token.Pos
	MaxPos token.Pos

	// Flag to track if the set has been initialized
	Initialized bool
}

// ensureInitialized initializes the set if it hasn't been initialized yet
// Safe to call on nil receiver - does nothing if receiver is nil.
func (s *IgnoreSet) ensureInitialized() {
	if s == nil {
		return
	}
	if !s.Initialized {
		s.Markers = make([]IgnoreMarker, 0)
		s.CodeIndex = make(map[string][]int)
		s.MinPos = token.NoPos
		s.MaxPos = token.NoPos
		s.Initialized = true
	}
}

// Add adds an annotation to the set and updates indices
// Safe to call on nil receiver - does nothing if receiver is nil.
func (s *IgnoreSet) Add(annotation IgnoreAnnotation) {
	s.ensureInitialized()
	if s == nil {
		return
	}

	// Convert interface to internal marker type
	marker := IgnoreMarker{
		Codes:    annotation.GetCodes(),
		StartPos: annotation.GetStartPos(),
		EndPos:   annotation.GetEndPos(),
	}

	// Add to markers list
	index := len(s.Markers)
	s.Markers = append(s.Markers, marker)

	// Update code index
	for _, code := range marker.Codes {
		s.CodeIndex[code] = append(s.CodeIndex[code], index)
	}

	// Update min/max positions
	if s.MinPos == token.NoPos || marker.StartPos < s.MinPos {
		s.MinPos = marker.StartPos
	}
	if s.MaxPos == token.NoPos || marker.EndPos > s.MaxPos {
		s.MaxPos = marker.EndPos
	}
}

// Contains checks if the given code is ignored at the specified position.
// Returns true if there's an @ignore annotation that covers the position.
//
// The method checks codes in hierarchical order using codes.GetCodesForCheck():
// - First checks "ALL" (universal ignore)
// - Then checks category prefix (e.g., "IMM" for "IMM01")
// - Finally checks the specific code (e.g., "IMM01")
//
// Example: for code "IMM01", it checks: "ALL", "IMM", "IMM01"
// Safe to call on nil receiver - returns false.
func (s *IgnoreSet) Contains(code string, pos token.Pos) bool {
	// Nil safety: return false if receiver is nil or uninitialized
	if s == nil || !s.Initialized {
		return false
	}

	// Quick range check: if pos is outside all markers, return false
	if s.MinPos == token.NoPos || pos < s.MinPos || pos > s.MaxPos {
		return false
	}

	// Check all codes in the hierarchy: ALL, category, specific code
	for checkCode := range codes.GetCodesForCheck(code) {
		indices, exists := s.CodeIndex[checkCode]
		if exists {
			for _, idx := range indices {
				marker := s.Markers[idx]
				if pos >= marker.StartPos && pos <= marker.EndPos {
					return true
				}
			}
		}
	}

	return false
}

// Len returns the number of markers in the set
// Safe to call on nil receiver - returns 0.
func (s *IgnoreSet) Len() int {
	if s == nil {
		return 0
	}
	return len(s.Markers)
}

// AddModuleIgnore adds ignore for entire module/flag from position 0 to max int
// This will ignore all violations for the specified codes everywhere
func (s *IgnoreSet) AddModuleIgnore(codes []string) {
	s.ensureInitialized()
	if s == nil {
		return
	}

	// Convert codes to uppercase
	upperCodes := make([]string, len(codes))
	for i, code := range codes {
		upperCodes[i] = strings.ToUpper(code)
	}

	// Create marker that covers entire range (0 to max int)
	marker := IgnoreMarker{
		Codes:    upperCodes,
		StartPos: 0,
		EndPos:   token.Pos(^uint(0) >> 1), // Max int for token.Pos
	}

	// Add to markers list
	index := len(s.Markers)
	s.Markers = append(s.Markers, marker)

	// Update code index
	for _, code := range marker.Codes {
		s.CodeIndex[code] = append(s.CodeIndex[code], index)
	}

	// Update min/max positions
	// Since this covers entire range, set extremes
	if s.MinPos == token.NoPos || 0 < s.MinPos {
		s.MinPos = 0
	}
	maxPos := token.Pos(^uint(0) >> 1)
	if s.MaxPos == token.NoPos || maxPos > s.MaxPos {
		s.MaxPos = maxPos
	}
}

// Empty returns true if the set contains no markers
// Safe to call on nil receiver - returns true.
func (s *IgnoreSet) Empty() bool {
	if s == nil {
		return true
	}
	return len(s.Markers) == 0
}
