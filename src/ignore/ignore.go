package ignore

import (
	"go/ast"
	"go/token"
	"regexp"
	"sort"
	"strings"

	"github.com/cloudflare/ahocorasick"
	"golang.org/x/tools/go/analysis"

	"goagreement/src/config"
	"goagreement/src/util"
)

// IgnoreAnnotation represents parsed @ignore CODE1, CODE2 annotation
// @immutable
// @implements &util.IgnoreAnnotation
// @constructor parseIgnoreAnnotation
type IgnoreAnnotation struct {
	// List of error codes to ignore (e.g., ["CODE1", "CODE2"])
	Codes []string

	// Start position of the ignore directive (comment position)
	StartPos token.Pos

	// End position where ignore directive ends
	// For now, this is the position of the next statement after the comment
	EndPos token.Pos
}

// GetCodes returns the list of error codes
func (a *IgnoreAnnotation) GetCodes() []string {
	return a.Codes
}

// GetStartPos returns the start position
func (a *IgnoreAnnotation) GetStartPos() token.Pos {
	return a.StartPos
}

// GetEndPos returns the end position
func (a *IgnoreAnnotation) GetEndPos() token.Pos {
	return a.EndPos
}

// IgnoreResult is the result type returned by IgnoreReader analyzer
// It contains all @ignore annotations found in the package using optimized IgnoreSet
// @immutable
type IgnoreResult struct {
	IgnoreSet *util.IgnoreSet
}

// Compile regex once
// Matches: @ignore CODE1, CODE2 or @ignore CODE1
var ignoreRegex = regexp.MustCompile(
	`^\s*//\s*@ignore(?:\s+(.+?))?\s*$`,
	//                            ^1
	// 1: comma-separated error codes (optional)
)

// parseIgnoreAnnotation parses string "@ignore CODE1, CODE2" or "@ignore CODE1"
// Returns nil if comment doesn't match @ignore pattern or has no codes
func parseIgnoreAnnotation(commentText string, startPos token.Pos, endPos token.Pos) *IgnoreAnnotation {
	match := ignoreRegex.FindStringSubmatch(commentText)
	if match == nil {
		return nil
	}

	// match[1] = "CODE1, CODE2" or ""
	codesStr := strings.TrimSpace(match[1])

	// If no codes provided, return nil (user must specify codes explicitly)
	if codesStr == "" {
		return nil
	}

	// Split by comma and trim each code
	var codes []string
	parts := strings.Split(codesStr, ",")
	for _, part := range parts {
		code := strings.TrimSpace(part)
		if code != "" {
			codes = append(codes, code)
		}
	}

	// If after trimming we have no codes, return nil
	if len(codes) == 0 {
		return nil
	}

	return &IgnoreAnnotation{
		Codes:    codes,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

// ignoreMatcher performs fast substring matching for @ignore comments.
// Note: Aho-Corasick is overkill for a single pattern, but used here
// for consistency with other annotation matchers in the codebase.
var ignoreMatcher = ahocorasick.NewStringMatcher([]string{
	"@ignore",
})

// ReadIgnoreAnnotations scans pass for @ignore annotations and returns IgnoreSet
// This function looks for @ignore comments and determines their scope
func ReadIgnoreAnnotations(pass *analysis.Pass) *util.IgnoreSet {
	ignoreSet := &util.IgnoreSet{}

	// Filter files based on configuration
	filesToScan := config.Global.FilterFiles(pass)

	for _, file := range filesToScan {
		// Scan all comment groups in the file
		for _, commentGroup := range file.Comments {
			for _, comment := range commentGroup.List {
				text := comment.Text

				// Micro-optimization: skip comments without @ignore
				if !ignoreMatcher.Contains([]byte(text)) {
					continue
				}

				// Parse @ignore annotation
				if strings.Contains(text, "@ignore") {
					startPos := comment.Pos()
					var endPos token.Pos

					// File-level annotation: comment before package declaration
					if startPos < file.Package {
						endPos = file.End()
					} else {
						// Check if this is an inline comment (on the same line as code)
						inlineStart, inlineEnd, isInline := findInlineNode(file, comment, pass.Fset)
						if isInline {
							// Inline comment: scope covers the entire line
							startPos = inlineStart
							endPos = inlineEnd
						} else {
							// Block comment: find the next node after comment
							endPos = findNextNodeAfterComment(file, startPos)

							// If no next node found, scope is just the comment itself
							if endPos == token.NoPos {
								endPos = comment.End()
							}
						}
					}

					annotation := parseIgnoreAnnotation(text, startPos, endPos)
					if annotation != nil {
						ignoreSet.Add(annotation)
					}
				}
			}
		}
	}

	return ignoreSet
}

// findInlineNode checks if comment is inline (on the same line as code).
// Returns (startPos, endPos, true) if inline, or (0, 0, false) if not inline.
// For inline comments, startPos is the beginning of the line, endPos is comment end.
// Example: var x int // @ignore CODE1
func findInlineNode(file *ast.File, comment *ast.Comment, fset *token.FileSet) (start token.Pos, end token.Pos, found bool) {
	commentPos := comment.Pos()
	commentLine := fset.Position(commentPos).Line

	// Binary search to find the declaration containing the comment
	idx := sort.Search(len(file.Decls), func(i int) bool {
		return file.Decls[i].End() > commentPos
	})

	// If no declaration found, not inline
	if idx >= len(file.Decls) {
		return 0, 0, false
	}

	decl := file.Decls[idx]

	// If comment is before this declaration, not inline
	if commentPos < decl.Pos() {
		return 0, 0, false
	}

	// Comment is inside declaration - check if there's code on same line before comment
	var hasCodeOnLine bool

	ast.Inspect(decl, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// Skip if node is after comment
		if n.Pos() >= commentPos {
			return false
		}

		nodeEndLine := fset.Position(n.End()).Line

		// Check if this node ends on the same line as the comment
		if nodeEndLine == commentLine {
			hasCodeOnLine = true
			return false // Found code, can stop
		}

		return true
	})

	// If no code on this line, it's not inline
	if !hasCodeOnLine {
		return 0, 0, false
	}

	// Return positions covering entire line (from line start to comment end)
	fileContent := fset.File(commentPos)
	if fileContent == nil {
		return 0, 0, false
	}

	lineStart := fileContent.LineStart(commentLine)
	return lineStart, comment.End(), true
}

// findNextNodeAfterComment finds the end position of the scope affected by @ignore comment.
// If comment is before a declaration (func, type, etc), returns the end of that declaration.
// If comment is inside a declaration, finds the next statement after comment.
// Returns token.NoPos if no node found after comment.
func findNextNodeAfterComment(file *ast.File, commentPos token.Pos) token.Pos {
	// Binary search to find the declaration that contains or follows the comment
	idx := sort.Search(len(file.Decls), func(i int) bool {
		return file.Decls[i].End() > commentPos
	})

	// If no declaration found after comment, return NoPos
	if idx >= len(file.Decls) {
		return token.NoPos
	}

	decl := file.Decls[idx]

	// If comment is before this declaration, @ignore applies to the entire declaration
	// Return the end of the declaration, not the beginning
	if commentPos < decl.Pos() {
		return decl.End()
	}

	// Comment is inside this declaration - find the next node after comment
	var nextPos = token.NoPos

	ast.Inspect(decl, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// Skip nodes that start before or at comment position
		if n.Pos() <= commentPos {
			return true
		}

		// Found a node after comment
		if nextPos == token.NoPos || n.Pos() < nextPos {
			nextPos = n.Pos()
			// Stop searching once we found the first node
			return false
		}

		return true
	})

	return nextPos
}
