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

// IsMeaningfulLine checks if a line of code is meaningful (not a comment or whitespace)
func IsMeaningfulLine(line string) bool {
	return !IsWhitespaceLine(line) && !IsCommentLine(line)
}
