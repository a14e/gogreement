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

func TestIgnoreSet_AddModuleIgnore(t *testing.T) {
	set := &IgnoreSet{}

	// Add module ignore for IMM01 and IMM02
	set.AddModuleIgnore([]string{"IMM01", "IMM02"})

	// Test that module ignores work regardless of position
	assert.True(t, set.Contains("IMM01", token.Pos(100)))   // Should be ignored
	assert.True(t, set.Contains("IMM02", token.Pos(200)))   // Should be ignored
	assert.False(t, set.Contains("IMM03", token.Pos(300)))  // Should not be ignored
	assert.False(t, set.Contains("CTOR01", token.Pos(400))) // Should not be ignored

	// Test hierarchical checking - IMM should match IMM01
	assert.True(t, set.Contains("IMM01", token.Pos(500))) // Should be ignored (direct match)

	// Add regular annotation and ensure both work
	ann := &mockAnnotation{
		codes:    []string{"CTOR01"},
		startPos: token.Pos(1000),
		endPos:   token.Pos(2000),
	}
	set.Add(ann)

	// Both module and regular ignores should work
	assert.True(t, set.Contains("IMM01", token.Pos(1500)))   // Module ignore
	assert.True(t, set.Contains("CTOR01", token.Pos(1500)))  // Regular ignore
	assert.False(t, set.Contains("CTOR02", token.Pos(1500))) // No ignore
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

func TestIgnoreSet_CategoryHierarchy(t *testing.T) {
	set := &IgnoreSet{}

	// Add annotation that ignores entire IMM category
	annCategory := &mockAnnotation{
		codes:    []string{"IMM"},
		startPos: token.Pos(100),
		endPos:   token.Pos(200),
	}
	set.Add(annCategory)

	// Add annotation that ignores specific code CTOR01
	annSpecific := &mockAnnotation{
		codes:    []string{"CTOR01"},
		startPos: token.Pos(300),
		endPos:   token.Pos(400),
	}
	set.Add(annSpecific)

	// IMM category should match all IMM codes in its range
	assert.True(t, set.Contains("IMM01", token.Pos(150)), "IMM category should match IMM01")
	assert.True(t, set.Contains("IMM02", token.Pos(150)), "IMM category should match IMM02")
	assert.True(t, set.Contains("IMM03", token.Pos(150)), "IMM category should match IMM03")
	assert.True(t, set.Contains("IMM04", token.Pos(150)), "IMM category should match IMM04")

	// IMM category should not match other categories
	assert.False(t, set.Contains("CTOR01", token.Pos(150)), "IMM category should not match CTOR01")
	assert.False(t, set.Contains("TONL01", token.Pos(150)), "IMM category should not match TONL01")

	// Specific code should match only itself in its range
	assert.True(t, set.Contains("CTOR01", token.Pos(350)), "CTOR01 should match in its range")
	assert.False(t, set.Contains("CTOR02", token.Pos(350)), "CTOR02 should not match in CTOR01 range")

	// Outside ranges should not match
	assert.False(t, set.Contains("IMM01", token.Pos(50)), "IMM01 should not match before range")
	assert.False(t, set.Contains("IMM01", token.Pos(250)), "IMM01 should not match between ranges")
}

func TestIgnoreSet_ALLOverridesEverything(t *testing.T) {
	set := &IgnoreSet{}

	// Add ALL annotation
	annAll := &mockAnnotation{
		codes:    []string{"ALL"},
		startPos: token.Pos(100),
		endPos:   token.Pos(200),
	}
	set.Add(annAll)

	// ALL should match all codes from all categories
	assert.True(t, set.Contains("IMM01", token.Pos(150)))
	assert.True(t, set.Contains("IMM02", token.Pos(150)))
	assert.True(t, set.Contains("CTOR01", token.Pos(150)))
	assert.True(t, set.Contains("CTOR02", token.Pos(150)))
	assert.True(t, set.Contains("TONL01", token.Pos(150)))
	assert.True(t, set.Contains("TONL02", token.Pos(150)))
	assert.True(t, set.Contains("TONL03", token.Pos(150)))

	// Even unknown codes
	assert.True(t, set.Contains("UNKNOWN99", token.Pos(150)))
}

func TestIgnoreSet_MultipleCategoriesAndSpecificCodes(t *testing.T) {
	set := &IgnoreSet{}

	// Ignore IMM category at 100-200
	ann1 := &mockAnnotation{
		codes:    []string{"IMM"},
		startPos: token.Pos(100),
		endPos:   token.Pos(200),
	}
	set.Add(ann1)

	// Ignore CTOR category at 300-400
	ann2 := &mockAnnotation{
		codes:    []string{"CTOR"},
		startPos: token.Pos(300),
		endPos:   token.Pos(400),
	}
	set.Add(ann2)

	// Ignore specific TONL01 at 500-600
	ann3 := &mockAnnotation{
		codes:    []string{"TONL01"},
		startPos: token.Pos(500),
		endPos:   token.Pos(600),
	}
	set.Add(ann3)

	// Test IMM range
	assert.True(t, set.Contains("IMM01", token.Pos(150)))
	assert.True(t, set.Contains("IMM02", token.Pos(150)))
	assert.False(t, set.Contains("CTOR01", token.Pos(150)))

	// Test CTOR range
	assert.False(t, set.Contains("IMM01", token.Pos(350)))
	assert.True(t, set.Contains("CTOR01", token.Pos(350)))
	assert.True(t, set.Contains("CTOR02", token.Pos(350)))

	// Test TONL01 specific range
	assert.False(t, set.Contains("IMM01", token.Pos(550)))
	assert.False(t, set.Contains("CTOR01", token.Pos(550)))
	assert.True(t, set.Contains("TONL01", token.Pos(550)))
	assert.False(t, set.Contains("TONL02", token.Pos(550)), "Only TONL01 should be ignored, not TONL02")
}

func TestIgnoreSet_OverlappingRanges(t *testing.T) {
	set := &IgnoreSet{}

	// Add IMM category at 100-300
	ann1 := &mockAnnotation{
		codes:    []string{"IMM"},
		startPos: token.Pos(100),
		endPos:   token.Pos(300),
	}
	set.Add(ann1)

	// Add specific IMM01 at 200-400 (overlaps with IMM category)
	ann2 := &mockAnnotation{
		codes:    []string{"IMM01"},
		startPos: token.Pos(200),
		endPos:   token.Pos(400),
	}
	set.Add(ann2)

	// In 100-200: IMM category covers IMM01
	assert.True(t, set.Contains("IMM01", token.Pos(150)))
	assert.True(t, set.Contains("IMM02", token.Pos(150)))

	// In 200-300: both IMM category and IMM01 specific cover IMM01
	assert.True(t, set.Contains("IMM01", token.Pos(250)))
	assert.True(t, set.Contains("IMM02", token.Pos(250)))

	// In 300-400: only IMM01 specific is active
	assert.True(t, set.Contains("IMM01", token.Pos(350)))
	assert.False(t, set.Contains("IMM02", token.Pos(350)), "IMM02 should not be covered after 300")
}
