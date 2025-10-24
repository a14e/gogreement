package util

import (
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockAnnotation implements IgnoreAnnotation interface for testing
type mockAnnotation struct {
	codes    []string
	startPos token.Pos
	endPos   token.Pos
}

func (m *mockAnnotation) GetCodes() []string {
	return m.codes
}

func (m *mockAnnotation) GetStartPos() token.Pos {
	return m.startPos
}

func (m *mockAnnotation) GetEndPos() token.Pos {
	return m.endPos
}

func TestIgnoreSet_ZeroValue(t *testing.T) {
	// Test that zero value IgnoreSet can be used
	set := &IgnoreSet{}
	assert.Equal(t, 0, set.Len())
	assert.True(t, set.Empty())

	// Add should initialize it
	ann := &mockAnnotation{
		codes:    []string{"CODE1"},
		startPos: token.Pos(10),
		endPos:   token.Pos(20),
	}
	set.Add(ann)
	assert.Equal(t, 1, set.Len())
	assert.False(t, set.Empty())
}

func TestIgnoreSet_Add(t *testing.T) {
	set := &IgnoreSet{}

	ann1 := &mockAnnotation{
		codes:    []string{"CODE1"},
		startPos: token.Pos(10),
		endPos:   token.Pos(20),
	}
	set.Add(ann1)

	assert.Equal(t, 1, set.Len())

	ann2 := &mockAnnotation{
		codes:    []string{"CODE2", "CODE3"},
		startPos: token.Pos(30),
		endPos:   token.Pos(40),
	}
	set.Add(ann2)

	assert.Equal(t, 2, set.Len())
}

func TestIgnoreSet_Contains(t *testing.T) {
	set := &IgnoreSet{}

	ann1 := &mockAnnotation{
		codes:    []string{"CODE1"},
		startPos: token.Pos(10),
		endPos:   token.Pos(20),
	}
	set.Add(ann1)

	ann2 := &mockAnnotation{
		codes:    []string{"CODE2", "CODE3"},
		startPos: token.Pos(30),
		endPos:   token.Pos(40),
	}
	set.Add(ann2)

	// Test position within first annotation
	assert.True(t, set.Contains("CODE1", token.Pos(15)))
	assert.True(t, set.Contains("CODE1", token.Pos(10))) // start boundary
	assert.True(t, set.Contains("CODE1", token.Pos(20))) // end boundary

	// Test position within second annotation
	assert.True(t, set.Contains("CODE2", token.Pos(35)))
	assert.True(t, set.Contains("CODE3", token.Pos(35)))

	// Test position outside annotations
	assert.False(t, set.Contains("CODE1", token.Pos(5)))  // before
	assert.False(t, set.Contains("CODE1", token.Pos(25))) // between
	assert.False(t, set.Contains("CODE2", token.Pos(25))) // between
	assert.False(t, set.Contains("CODE1", token.Pos(45))) // after

	// Test non-existent code
	assert.False(t, set.Contains("CODE999", token.Pos(15)))
}

func TestIgnoreSet_Contains_EmptySet(t *testing.T) {
	set := &IgnoreSet{}

	assert.False(t, set.Contains("CODE1", token.Pos(10)))
}

func TestIgnoreSet_Markers(t *testing.T) {
	set := &IgnoreSet{}

	ann1 := &mockAnnotation{
		codes:    []string{"CODE1"},
		startPos: token.Pos(10),
		endPos:   token.Pos(20),
	}
	set.Add(ann1)

	ann2 := &mockAnnotation{
		codes:    []string{"CODE2"},
		startPos: token.Pos(30),
		endPos:   token.Pos(40),
	}
	set.Add(ann2)

	assert.Len(t, set.Markers, 2)
	assert.Equal(t, []string{"CODE1"}, set.Markers[0].Codes)
	assert.Equal(t, token.Pos(10), set.Markers[0].StartPos)
	assert.Equal(t, token.Pos(20), set.Markers[0].EndPos)
	assert.Equal(t, []string{"CODE2"}, set.Markers[1].Codes)
	assert.Equal(t, token.Pos(30), set.Markers[1].StartPos)
	assert.Equal(t, token.Pos(40), set.Markers[1].EndPos)
}

func TestIgnoreSet_MinMaxPositions(t *testing.T) {
	set := &IgnoreSet{}

	// Add annotation with positions 50-100
	ann1 := &mockAnnotation{
		codes:    []string{"CODE1"},
		startPos: token.Pos(50),
		endPos:   token.Pos(100),
	}
	set.Add(ann1)

	// Add annotation with positions 10-30 (should update minPos)
	ann2 := &mockAnnotation{
		codes:    []string{"CODE2"},
		startPos: token.Pos(10),
		endPos:   token.Pos(30),
	}
	set.Add(ann2)

	// Add annotation with positions 80-200 (should update maxPos)
	ann3 := &mockAnnotation{
		codes:    []string{"CODE3"},
		startPos: token.Pos(80),
		endPos:   token.Pos(200),
	}
	set.Add(ann3)

	// Test that positions outside min/max return false quickly
	assert.False(t, set.Contains("CODE1", token.Pos(5)))   // before minPos (10)
	assert.False(t, set.Contains("CODE1", token.Pos(201))) // after maxPos (200)

	// Test positions within range
	assert.True(t, set.Contains("CODE2", token.Pos(20)))
	assert.True(t, set.Contains("CODE3", token.Pos(150)))
}

func TestIgnoreSet_CodeIndex(t *testing.T) {
	set := &IgnoreSet{}

	// Add multiple annotations with same code
	ann1 := &mockAnnotation{
		codes:    []string{"COMMON"},
		startPos: token.Pos(10),
		endPos:   token.Pos(20),
	}
	set.Add(ann1)

	ann2 := &mockAnnotation{
		codes:    []string{"COMMON"},
		startPos: token.Pos(50),
		endPos:   token.Pos(60),
	}
	set.Add(ann2)

	// Should find COMMON at both positions
	assert.True(t, set.Contains("COMMON", token.Pos(15)))
	assert.True(t, set.Contains("COMMON", token.Pos(55)))

	// But not in between
	assert.False(t, set.Contains("COMMON", token.Pos(30)))
}

func TestIgnoreSet_ALL_Code(t *testing.T) {
	set := &IgnoreSet{}

	// Add ALL annotation
	annAll := &mockAnnotation{
		codes:    []string{"ALL"},
		startPos: token.Pos(100),
		endPos:   token.Pos(200),
	}
	set.Add(annAll)

	// Add specific code annotation
	annSpecific := &mockAnnotation{
		codes:    []string{"SPECIFIC"},
		startPos: token.Pos(300),
		endPos:   token.Pos(400),
	}
	set.Add(annSpecific)

	// ALL should match any code in its range
	assert.True(t, set.Contains("ANYCODE", token.Pos(150)), "ALL should match any code")
	assert.True(t, set.Contains("CODE1", token.Pos(150)), "ALL should match CODE1")
	assert.True(t, set.Contains("CODE2", token.Pos(150)), "ALL should match CODE2")

	// ALL should not match outside its range
	assert.False(t, set.Contains("ANYCODE", token.Pos(50)), "ALL should not match before range")
	assert.False(t, set.Contains("ANYCODE", token.Pos(250)), "ALL should not match after ALL range but before SPECIFIC range")

	// Specific code should still work in its range
	assert.True(t, set.Contains("SPECIFIC", token.Pos(350)))

	// ALL should match SPECIFIC code in ALL's range (even though SPECIFIC is not defined there)
	assert.True(t, set.Contains("SPECIFIC", token.Pos(150)))

	// Checking for ALL explicitly
	assert.True(t, set.Contains("ALL", token.Pos(150)))
	assert.False(t, set.Contains("ALL", token.Pos(350)))
}

func TestIgnoreSet_MultipleALL(t *testing.T) {
	set := &IgnoreSet{}

	// Add multiple ALL annotations at different ranges
	ann1 := &mockAnnotation{
		codes:    []string{"ALL"},
		startPos: token.Pos(10),
		endPos:   token.Pos(50),
	}
	set.Add(ann1)

	ann2 := &mockAnnotation{
		codes:    []string{"ALL"},
		startPos: token.Pos(100),
		endPos:   token.Pos(150),
	}
	set.Add(ann2)

	// Both ranges should ignore all codes
	assert.True(t, set.Contains("CODE1", token.Pos(30)))
	assert.True(t, set.Contains("CODE2", token.Pos(30)))
	assert.True(t, set.Contains("CODE1", token.Pos(120)))
	assert.True(t, set.Contains("CODE2", token.Pos(120)))

	// Between ranges should not ignore
	assert.False(t, set.Contains("CODE1", token.Pos(75)))
}

func TestIgnoreSet_ALLWithSpecificCodes(t *testing.T) {
	set := &IgnoreSet{}

	// Add annotation with both ALL and specific codes
	ann := &mockAnnotation{
		codes:    []string{"ALL", "CODE1", "CODE2"},
		startPos: token.Pos(100),
		endPos:   token.Pos(200),
	}
	set.Add(ann)

	// All codes should match
	assert.True(t, set.Contains("ALL", token.Pos(150)))
	assert.True(t, set.Contains("CODE1", token.Pos(150)))
	assert.True(t, set.Contains("CODE2", token.Pos(150)))
	assert.True(t, set.Contains("ANYCODE", token.Pos(150)))
}

func TestIgnoreSet_Empty(t *testing.T) {
	set := &IgnoreSet{}

	assert.True(t, set.Empty())

	ann := &mockAnnotation{
		codes:    []string{"CODE1"},
		startPos: token.Pos(10),
		endPos:   token.Pos(20),
	}
	set.Add(ann)
	assert.False(t, set.Empty())

	ann2 := &mockAnnotation{
		codes:    []string{"CODE2"},
		startPos: token.Pos(30),
		endPos:   token.Pos(40),
	}
	set.Add(ann2)
	assert.False(t, set.Empty())
}
