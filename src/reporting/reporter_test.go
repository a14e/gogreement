package reporting

import (
	"go/token"
	"strings"
	"testing"

	"github.com/a14e/gogreement/src/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
)

func TestGetFileLinesLongLine(t *testing.T) {
	// A single line longer than bufio.Scanner's default 64KB token size must
	// not truncate the file (which would drop later lines' source snippets).
	longLine := strings.Repeat("x", 70*1024)
	content := "package p\n" + longLine + "\nlast := 1\n"

	pass := &analysis.Pass{
		ReadFile: func(string) ([]byte, error) { return []byte(content), nil },
	}
	r := NewReporter(pass, nil)

	lines := r.getFileLines("fake.go")
	require.GreaterOrEqual(t, len(lines), 3, "all lines must be read, not truncated at the long line")
	assert.Equal(t, longLine, lines[1], "the >64KB line must be captured intact")
	assert.Equal(t, "last := 1", lines[2], "lines after the long line must still be present")
}

// MockViolation implements Violation interface for testing
type MockViolation struct {
	code    string
	pos     token.Pos
	message string
}

func (m MockViolation) GetCode() string {
	return m.code
}

func (m MockViolation) GetPos() token.Pos {
	return m.pos
}

func (m MockViolation) GetMessage() string {
	return m.message
}

func TestNewReporter(t *testing.T) {
	pass := &analysis.Pass{}
	ignoreSet := &util.IgnoreSet{}

	reporter := NewReporter(pass, ignoreSet)

	assert.NotNil(t, reporter)
	assert.Equal(t, pass, reporter.pass)
	assert.Equal(t, ignoreSet, reporter.ignoreSet)
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxLen   int
		pos      int
		expected string
	}{
		{
			name:     "Short string unchanged",
			s:        "short",
			maxLen:   10,
			pos:      3,
			expected: "short",
		},
		{
			name:     "Position fits in first part",
			s:        "short string",
			maxLen:   10,
			pos:      5,
			expected: "short s...",
		},
		{
			name:     "Position near end",
			s:        "very long string with position at the end",
			maxLen:   20,
			pos:      40,
			expected: "...sition at the end",
		},
		{
			name:     "Position in middle",
			s:        "very long string with position in the middle here",
			maxLen:   20,
			pos:      25,
			expected: "... with position in...",
		},
		{
			name:     "Position at start",
			s:        "long string starting here",
			maxLen:   15,
			pos:      1,
			expected: "long string ...",
		},
		{
			name:     "MaxLineLength constant",
			s:        "this is a test for the max line length constant to see if it works correctly",
			maxLen:   MaxLineLength,
			pos:      30,
			expected: "this is a test for the max line length constant to see if it works correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.s, tt.maxLen, tt.pos)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateDisplayColumn(t *testing.T) {
	tests := []struct {
		name         string
		originalLine string
		originalPos  int
		maxLen       int
		expected     int
	}{
		{
			name:         "Short line unchanged",
			originalLine: "short line",
			originalPos:  5,
			maxLen:       20,
			expected:     5,
		},
		{
			name:         "Position fits in first part",
			originalLine: "short line here",
			originalPos:  10,
			maxLen:       15,
			expected:     10,
		},
		{
			name:         "Position near end",
			originalLine: "very long string with position at the end",
			originalPos:  40,
			maxLen:       20,
			expected:     19, // 4 for "..." + position in end part
		},
		{
			name:         "Position in middle",
			originalLine: "very long string with position in the middle here",
			originalPos:  25,
			maxLen:       20,
			expected:     12, // 4 for "..." + middle position
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateDisplayColumn(tt.originalLine, tt.originalPos, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReportViolations(t *testing.T) {
	// This is a basic test - full testing would require setting up a complete analysis.Pass
	// with proper file system and token positions
	t.Run("Empty violations list", func(t *testing.T) {
		pass := &analysis.Pass{}
		ignoreSet := &util.IgnoreSet{}
		violations := []Violation{}

		reporter := NewReporter(pass, ignoreSet)
		// Should not panic
		reporter.ReportViolations(violations)
	})
}
