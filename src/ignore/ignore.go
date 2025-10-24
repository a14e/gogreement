package ignore

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"github.com/cloudflare/ahocorasick"
	"golang.org/x/tools/go/analysis"

	"goagreement/src/config"
	"goagreement/src/util"
)

// IgnoreAnnotation represents parsed @ignore CODE1, CODE2 annotation
// @immutable
// @implements &util.IgnoreAnnotation
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
		// Lazy initialization: build a comment-to-statement map only if we find at least one @ignore
		var commentToNextStmt map[token.Pos]token.Pos

		// Determine the first declaration position for file-level annotations
		var firstDeclPos = token.NoPos
		if len(file.Decls) > 0 {
			firstDeclPos = file.Decls[0].Pos()
		}

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
					// Lazy initialization: build map only on first @ignore found
					if commentToNextStmt == nil {
						commentToNextStmt = buildCommentToStmtMap(file, pass.Fset)
					}

					startPos := comment.Pos()
					endPos := commentToNextStmt[startPos]

					// If we couldn't find next statement, check if it's a file-level annotation
					if endPos == token.NoPos {
						// If comment is before first declaration, it's file-level
						if firstDeclPos != token.NoPos && startPos < firstDeclPos {
							endPos = file.End()
						} else {
							endPos = comment.End()
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

// buildCommentToStmtMap builds a mapping from comment position to the position
// of the next statement that follows it. This is used to determine the scope
// of @ignore directives.
func buildCommentToStmtMap(file *ast.File, fset *token.FileSet) map[token.Pos]token.Pos {
	result := make(map[token.Pos]token.Pos)

	// Walk through all declarations and statements to find what comes after each comment
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			// Map function comments to function body start
			if node.Doc != nil && node.Body != nil {
				for _, comment := range node.Doc.List {
					result[comment.Pos()] = node.Body.Lbrace
				}
			}
			// Process function body statements
			if node.Body != nil {
				mapCommentsInStmtList(node.Body.List, fset, result)
			}

		case *ast.GenDecl:
			// Map declaration comments to declaration start
			if node.Doc != nil {
				for _, comment := range node.Doc.List {
					result[comment.Pos()] = node.Pos()
				}
			}
		}
		return true
	})

	return result
}

// mapCommentsInStmtList maps comments in a statement list to their following statements
func mapCommentsInStmtList(stmts []ast.Stmt, fset *token.FileSet, result map[token.Pos]token.Pos) {
	for i, stmt := range stmts {
		// Check if this statement has a comment before it
		// We'll look at the previous line to see if there's a comment

		// For now, we'll use a simpler approach: map any statement position
		// to the next statement position (if exists), or to its own end
		if i < len(stmts)-1 {
			// There's a next statement
			nextStmt := stmts[i+1]
			// Any comment on this statement's line maps to next statement
			result[stmt.Pos()] = nextStmt.Pos()
		} else {
			// Last statement - map to its end
			result[stmt.Pos()] = stmt.End()
		}

		// Recursively handle nested statements
		ast.Inspect(stmt, func(n ast.Node) bool {
			if blockStmt, ok := n.(*ast.BlockStmt); ok {
				mapCommentsInStmtList(blockStmt.List, fset, result)
			}
			return true
		})
	}
}
