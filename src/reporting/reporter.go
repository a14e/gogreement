package reporting

import (
	"bufio"
	"fmt"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/codes"
	"github.com/a14e/gogreement/src/util"
)

const (
	// MaxLineLength is the maximum length of a line displayed in error messages
	MaxLineLength = 200
)

// Violation represents a generic violation interface
// All violation types should implement this interface
type Violation interface {
	// GetCode returns the error code for this violation
	GetCode() string

	// GetPos returns the position of the violation
	GetPos() token.Pos

	// GetMessage returns the main error message without formatting
	GetMessage() string
}

// Reporter handles violation reporting with pretty formatting
type Reporter struct {
	pass      *analysis.Pass
	ignoreSet *util.IgnoreSet
	lineCache map[string][]string // filename -> cached lines
}

func NewReporter(pass *analysis.Pass, ignoreSet *util.IgnoreSet) *Reporter {
	return &Reporter{
		pass:      pass,
		ignoreSet: ignoreSet,
		lineCache: make(map[string][]string),
	}
}

func (r *Reporter) ReportViolation(violation Violation) {
	if r.ignoreSet.Contains(violation.GetCode(), violation.GetPos()) {
		return
	}

	r.pass.Report(analysis.Diagnostic{
		Pos:     violation.GetPos(),
		Message: r.formatPrettyError(violation),
	})
}

func (r *Reporter) ReportViolations(violations []Violation) {
	for _, violation := range violations {
		r.ReportViolation(violation)
	}
}

// formatPrettyError formats an error with pretty borders and help
func (r *Reporter) formatPrettyError(violation Violation) string {
	position := r.pass.Fset.Position(violation.GetPos())

	lines := r.readSourceLines(position.Filename, position.Line, 2, 1) // 2 lines before, 1 line after

	var builder strings.Builder

	// Error header with code in brackets
	builder.WriteString("error: ")
	builder.WriteString("[")
	builder.WriteString(violation.GetCode())
	builder.WriteString("] ")
	builder.WriteString(violation.GetMessage())
	builder.WriteString("\n")

	// Source code context
	if len(lines.content) > 0 {
		// Calculate max line number width for proper alignment
		maxLineNum := 0
		for _, num := range lines.lineNumbers {
			if num > maxLineNum {
				maxLineNum = num
			}
		}
		lineNumWidth := len(fmt.Sprintf("%d", maxLineNum))

		// Add top border with proper alignment
		builder.WriteString(strings.Repeat(" ", lineNumWidth))
		builder.WriteString(" |\n")

		// Display lines with context
		for i, line := range lines.content {
			lineNum := lines.lineNumbers[i]
			truncatedLine := truncateString(line, MaxLineLength, position.Column)
			builder.WriteString(fmt.Sprintf("%*d | ", lineNumWidth, lineNum))
			builder.WriteString(truncatedLine)
			builder.WriteString("\n")

			// Add pointer under the error line
			if lineNum == position.Line {
				builder.WriteString(strings.Repeat(" ", lineNumWidth))
				builder.WriteString(" | ")

				// Calculate column position in truncated line
				displayColumn := calculateDisplayColumn(line, position.Column, MaxLineLength)

				// Add spaces to align the pointer
				for i := 1; i < displayColumn; i++ {
					if i-1 < len(truncatedLine) && truncatedLine[i-1] == '\t' {
						builder.WriteString("\t")
					} else {
						builder.WriteString(" ")
					}
				}
				builder.WriteString("^\n")
			}
		}

		// Help section with documentation link
		builder.WriteString(strings.Repeat(" ", lineNumWidth))
		builder.WriteString(" |\n")
		builder.WriteString("   = help: ")
		builder.WriteString(codes.GetDocumentationURL(violation.GetCode()))
		builder.WriteString("\n")
	}

	return builder.String()
}

// sourceLines represents a set of source lines with their numbers
type sourceLines struct {
	content     []string
	lineNumbers []int
}

// getFileLines reads and caches all lines from a file
func (r *Reporter) getFileLines(filename string) []string {
	if lines, exists := r.lineCache[filename]; exists {
		return lines
	}

	content, err := r.pass.ReadFile(filename)
	if err != nil {
		return nil
	}

	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	r.lineCache[filename] = lines
	return lines
}

// readSourceLines reads source file lines around the given line number using analysis.Pass
// Uses cached lines to avoid repeated file parsing
func (r *Reporter) readSourceLines(filename string, lineNum, before, after int) sourceLines {
	lines := r.getFileLines(filename)
	if lines == nil {
		return sourceLines{}
	}

	start := lineNum - before - 1 // Convert to 0-based index
	if start < 0 {
		start = 0
	}

	end := lineNum + after - 1 // Convert to 0-based index
	if end >= len(lines) {
		end = len(lines) - 1
	}

	if start >= len(lines) {
		return sourceLines{}
	}

	var result sourceLines
	for i := start; i <= end; i++ {
		result.content = append(result.content, lines[i])
		result.lineNumbers = append(result.lineNumbers, i+1) // Convert back to 1-based
	}

	return result
}

// truncateString truncates a string to the specified length, ensuring position is included
func truncateString(s string, maxLen int, pos int) string {
	if len(s) <= maxLen {
		return s
	}

	if maxLen <= 3 {
		return s[:maxLen]
	}

	// Convert pos to 0-based index
	pos0 := pos - 1
	if pos0 < 0 {
		pos0 = 0
	}
	if pos0 >= len(s) {
		pos0 = len(s) - 1
	}

	// If position fits in the first part, truncate from end
	if pos0 <= maxLen-3 {
		return s[:maxLen-3] + "..."
	}

	// If position is near the end, truncate from beginning
	if pos0 >= len(s)-maxLen+3 {
		return "..." + s[len(s)-maxLen+3:]
	}

	// Position is in the middle, truncate from both sides
	// Try to keep roughly equal characters before and after position
	before := (maxLen - 3) / 2
	after := (maxLen - 3) - before

	start := pos0 - before
	if start < 0 {
		start = 0
	}

	end := pos0 + after
	if end > len(s) {
		end = len(s)
	}

	return "..." + s[start:end] + "..."
}

// calculateDisplayColumn calculates the column position in the truncated string
func calculateDisplayColumn(originalLine string, originalPos, maxLen int) int {
	if len(originalLine) <= maxLen {
		return originalPos
	}

	// Convert to 0-based
	pos0 := originalPos - 1
	if pos0 < 0 {
		return 1
	}
	if pos0 >= len(originalLine) {
		pos0 = len(originalLine) - 1
	}

	// If position fits in first part
	if pos0 <= maxLen-3 {
		return originalPos
	}

	// If position is near end
	if pos0 >= len(originalLine)-maxLen+3 {
		return 4 + (pos0 - (len(originalLine) - maxLen + 3)) // 4 for "..."
	}

	// Position is in middle
	before := (maxLen - 3) / 2
	return 4 + before // 4 for "..." + position in middle section
}
