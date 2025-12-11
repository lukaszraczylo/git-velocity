package diff

import (
	"strings"
)

// IsCommentLine checks if a line is a code comment (should not count as meaningful contribution)
func IsCommentLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true // Empty lines don't count
	}

	// Common comment patterns across languages
	commentPrefixes := []string{
		"//",     // C, C++, Java, Go, JS, TS, Swift, Kotlin, etc.
		"#",      // Python, Ruby, Shell, YAML, Perl, etc.
		"/*",     // C-style block comment start
		"*/",     // C-style block comment end
		"*",      // C-style block comment continuation
		"<!--",   // HTML/XML comment
		"-->",    // HTML/XML comment end
		"--",     // SQL, Lua, Haskell
		";",      // Assembly, Lisp, INI files
		"'",      // VB comment
		"\"\"\"", // Python docstring
		"'''",    // Python docstring
	}

	for _, prefix := range commentPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}

	return false
}

// IsWhitespaceLine checks if a line contains only whitespace characters
func IsWhitespaceLine(line string) bool {
	return strings.TrimSpace(line) == ""
}

// IsDocumentationFile checks if a file is documentation-only
func IsDocumentationFile(filename string) bool {
	// Documentation file extensions and patterns
	docPatterns := []string{
		".md", ".markdown", ".rst", ".txt", ".adoc",
		"README", "CHANGELOG", "LICENSE", "CONTRIBUTING",
		"docs/", "documentation/", "/doc/",
	}

	lowerFilename := strings.ToLower(filename)
	for _, pattern := range docPatterns {
		if strings.Contains(lowerFilename, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// PatchStats holds the results of analyzing a diff patch
type PatchStats struct {
	TotalAdditions      int
	TotalDeletions      int
	MeaningfulAdditions int
	MeaningfulDeletions int
	CommentAdditions    int
	CommentDeletions    int
	WhitespaceAdditions int
	WhitespaceDeletions int
}

// AnalyzePatch analyzes a unified diff patch and returns both raw and meaningful line counts.
// It parses diff hunks and categorizes each changed line as meaningful, comment, or whitespace.
func AnalyzePatch(patch string) PatchStats {
	stats := PatchStats{}

	lines := strings.Split(patch, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		// Check if this is an addition or deletion line
		isAddition := strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++")
		isDeletion := strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---")

		if !isAddition && !isDeletion {
			continue // Context line or header
		}

		// Remove the diff prefix to get actual content
		content := line[1:]

		// Categorize the line
		if IsWhitespaceLine(content) {
			if isAddition {
				stats.TotalAdditions++
				stats.WhitespaceAdditions++
			} else {
				stats.TotalDeletions++
				stats.WhitespaceDeletions++
			}
		} else if IsCommentLine(content) {
			if isAddition {
				stats.TotalAdditions++
				stats.CommentAdditions++
			} else {
				stats.TotalDeletions++
				stats.CommentDeletions++
			}
		} else {
			// Meaningful code line
			if isAddition {
				stats.TotalAdditions++
				stats.MeaningfulAdditions++
			} else {
				stats.TotalDeletions++
				stats.MeaningfulDeletions++
			}
		}
	}

	return stats
}

// AnalyzePatchSimple returns just the meaningful additions and deletions
func AnalyzePatchSimple(patch string) (meaningfulAdds, meaningfulDels int) {
	stats := AnalyzePatch(patch)
	return stats.MeaningfulAdditions, stats.MeaningfulDeletions
}

// IsMeaningfulLine checks if a line of code is meaningful (not a comment or whitespace)
func IsMeaningfulLine(line string) bool {
	return !IsWhitespaceLine(line) && !IsCommentLine(line)
}
