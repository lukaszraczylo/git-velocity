package diff

import (
	"strings"
)

// IsCommentLine checks if a line is a code comment (should not count as meaningful contribution)
// Note: Empty/whitespace lines are NOT comments - use IsWhitespaceLine for those.
func IsCommentLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false // Empty lines are whitespace, not comments
	}

	// Common comment patterns across languages
	// Order matters for overlapping prefixes (e.g., "///" before "//")
	commentPrefixes := []string{
		"///",    // Rust/Swift/C# doc comments
		"//",     // C, C++, Java, Go, JS, TS, Swift, Kotlin, etc.
		"#",      // Python, Ruby, Shell, YAML, Perl, etc.
		"/**",    // JSDoc/JavaDoc block start
		"/*",     // C-style block comment start
		"*/",     // C-style block comment end
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

	// C-style block comment continuation: line starts with * followed by space or end of line
	// This avoids false positives like "*ptr = value" (pointer dereference)
	if strings.HasPrefix(trimmed, "*") {
		if len(trimmed) == 1 {
			return true // Just "*" alone
		}
		// Must be followed by whitespace or common comment characters, not alphanumeric
		nextChar := trimmed[1]
		if nextChar == ' ' || nextChar == '\t' || nextChar == '/' {
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

// IsDocCommentLine checks if a line is a documentation comment (JSDoc, JavaDoc, Rust doc, etc.)
// These are comments specifically meant to document code, as opposed to regular comments.
func IsDocCommentLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	// Documentation comment patterns
	docPrefixes := []string{
		"///",    // Rust, Swift, C# doc comments
		"//!",    // Rust inner doc comments
		"/**",    // JSDoc, JavaDoc block start
		"\"\"\"", // Python docstring
		"'''",    // Python docstring
	}

	for _, prefix := range docPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}

	// JSDoc/JavaDoc continuation lines with annotations (@param, @return, etc.)
	if strings.HasPrefix(trimmed, "* @") || strings.HasPrefix(trimmed, "*  @") {
		return true
	}

	// Check for common doc annotations at the start of a comment
	if strings.HasPrefix(trimmed, "// @") || strings.HasPrefix(trimmed, "# @") {
		return true
	}

	return false
}

// IsCommentedOutCode attempts to detect if a comment line contains commented-out code
// rather than an actual comment. This is a heuristic and may have false positives/negatives.
func IsCommentedOutCode(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	// Remove comment prefix to get the content
	var content string
	commentPrefixes := []string{"///", "//", "#", "/*", "--", ";"}
	for _, prefix := range commentPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			content = strings.TrimSpace(trimmed[len(prefix):])
			break
		}
	}

	if content == "" {
		return false
	}

	// Heuristics for detecting commented-out code:
	// 1. Ends with common code patterns
	codeEndings := []string{";", "{", "}", ")", ",", ":", "=>", "->"}
	for _, ending := range codeEndings {
		if strings.HasSuffix(content, ending) {
			return true
		}
	}

	// 2. Starts with common code keywords
	codeKeywords := []string{
		"if ", "else ", "for ", "while ", "switch ", "case ", "return ", "break", "continue",
		"const ", "let ", "var ", "func ", "function ", "def ", "class ", "struct ", "type ",
		"import ", "from ", "package ", "public ", "private ", "protected ", "static ",
		"async ", "await ", "try ", "catch ", "throw ", "raise ",
	}
	contentLower := strings.ToLower(content)
	for _, keyword := range codeKeywords {
		if strings.HasPrefix(contentLower, keyword) {
			return true
		}
	}

	// 3. Contains assignment operators
	if strings.Contains(content, " = ") || strings.Contains(content, " := ") ||
		strings.Contains(content, " == ") || strings.Contains(content, " != ") {
		return true
	}

	return false
}

// IsRenameOrMove checks if a file change represents a rename or move operation
// rather than actual content modification. A rename/move is detected when both
// the source (fromName) and destination (toName) paths exist and differ.
func IsRenameOrMove(fromName, toName string) bool {
	return fromName != "" && toName != "" && fromName != toName
}
