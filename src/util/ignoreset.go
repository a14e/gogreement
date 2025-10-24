package util

import (
	"go/token"
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
func (s *IgnoreSet) ensureInitialized() {
	if !s.Initialized {
		s.Markers = make([]IgnoreMarker, 0)
		s.CodeIndex = make(map[string][]int)
		s.MinPos = token.NoPos
		s.MaxPos = token.NoPos
		s.Initialized = true
	}
}

// Add adds an annotation to the set and updates indices
func (s *IgnoreSet) Add(annotation IgnoreAnnotation) {
	s.ensureInitialized()

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

// Contains checks if the given code is ignored at the specified position
// Returns true if there's an @ignore annotation with this code that covers the position
// Special code "ALL" matches any position covered by an ALL annotation
func (s *IgnoreSet) Contains(code string, pos token.Pos) bool {
	// Quick range check: if pos is outside all markers, return false
	if s.MinPos == token.NoPos || pos < s.MinPos || pos > s.MaxPos {
		return false
	}

	// Check if specific code is ignored at this position
	indices, exists := s.CodeIndex[code]
	if exists {
		for _, idx := range indices {
			marker := s.Markers[idx]
			if pos >= marker.StartPos && pos <= marker.EndPos {
				return true
			}
		}
	}

	// Check if ALL code is present at this position (unless we're already checking for ALL)
	if code != "ALL" {
		allIndices, allExists := s.CodeIndex["ALL"]
		if allExists {
			for _, idx := range allIndices {
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
func (s *IgnoreSet) Len() int {
	return len(s.Markers)
}
