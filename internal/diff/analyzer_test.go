package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCommentLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		// Empty and whitespace
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"tab only", "\t", true},
		{"mixed whitespace", "  \t  ", true},

		// C-style comments (Go, Java, JS, C++, etc.)
		{"C single line comment", "// this is a comment", true},
		{"C single line with leading space", "  // this is a comment", true},
		{"C block start", "/* block comment", true},
		{"C block end", "*/", true},
		{"C block continuation", "* continuation", true},
		{"C block continuation with space", "  * continuation", true},

		// Python/Shell comments
		{"Python comment", "# python comment", true},
		{"Shell comment", "#!/bin/bash", true},
		{"Python comment with space", "  # comment", true},

		// Python docstrings
		{"Python docstring double", "\"\"\"docstring", true},
		{"Python docstring single", "'''docstring", true},

		// SQL/Lua/Haskell comments
		{"SQL comment", "-- SQL comment", true},

		// Assembly/Lisp/INI comments
		{"Assembly comment", "; assembly comment", true},
		{"INI comment", "; ini comment", true},

		// VB comments
		{"VB comment", "' VB comment", true},

		// HTML/XML comments
		{"HTML comment start", "<!-- html comment", true},
		{"HTML comment end", "-->", true},

		// Actual code - NOT comments
		{"Go code", "func main() {", false},
		{"Python code", "def main():", false},
		{"JS code", "const x = 5;", false},
		{"Variable assignment", "x = 10", false},
		{"Return statement", "return nil", false},
		{"Import statement", "import fmt", false},
		{"Package declaration", "package main", false},
		{"Struct field", "Name string", false},
		{"Function call", "fmt.Println(x)", false},
		{"String with slash", `"http://example.com"`, false},
		{"Code after whitespace", "    x := 5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCommentLine(tt.line)
			assert.Equal(t, tt.expected, result, "IsCommentLine(%q)", tt.line)
		})
	}
}

func TestIsWhitespaceLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"empty string", "", true},
		{"single space", " ", true},
		{"multiple spaces", "     ", true},
		{"single tab", "\t", true},
		{"multiple tabs", "\t\t\t", true},
		{"mixed whitespace", "  \t  \t  ", true},
		{"newline only", "\n", true},
		{"carriage return", "\r", true},
		{"code line", "x := 5", false},
		{"code with leading whitespace", "    x := 5", false},
		{"comment line", "// comment", false},
		{"single character", "x", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWhitespaceLine(tt.line)
			assert.Equal(t, tt.expected, result, "IsWhitespaceLine(%q)", tt.line)
		})
	}
}

func TestIsDocumentationFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		// Documentation files
		{"readme markdown", "README.md", true},
		{"readme uppercase", "README", true},
		{"readme lowercase", "readme.md", true},
		{"changelog", "CHANGELOG.md", true},
		{"license", "LICENSE", true},
		{"license txt", "LICENSE.txt", true},
		{"contributing", "CONTRIBUTING.md", true},
		{"markdown file", "docs.md", true},
		{"rst file", "index.rst", true},
		{"txt file", "notes.txt", true},
		{"adoc file", "guide.adoc", true},
		{"docs directory", "docs/api.md", true},
		{"documentation directory", "documentation/guide.md", true},
		{"doc directory", "/doc/api.md", true},

		// Code files - NOT documentation
		{"go file", "main.go", false},
		{"python file", "app.py", false},
		{"js file", "index.js", false},
		{"ts file", "app.ts", false},
		{"java file", "App.java", false},
		{"c file", "main.c", false},
		{"cpp file", "main.cpp", false},
		{"rust file", "main.rs", false},
		{"yaml file", "config.yaml", false},
		{"json file", "package.json", false},
		{"html file", "index.html", false},
		{"css file", "style.css", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDocumentationFile(tt.filename)
			assert.Equal(t, tt.expected, result, "IsDocumentationFile(%q)", tt.filename)
		})
	}
}

func TestAnalyzePatch(t *testing.T) {
	tests := []struct {
		name     string
		patch    string
		expected PatchStats
	}{
		{
			name: "simple additions",
			patch: `@@ -1,3 +1,5 @@
 context line
+func main() {
+    x := 5
+}`,
			expected: PatchStats{
				TotalAdditions:      3,
				MeaningfulAdditions: 3,
			},
		},
		{
			name: "simple deletions",
			patch: `@@ -1,5 +1,3 @@
 context line
-func main() {
-    x := 5
-}`,
			expected: PatchStats{
				TotalDeletions:      3,
				MeaningfulDeletions: 3,
			},
		},
		{
			name: "mixed additions and deletions",
			patch: `@@ -1,3 +1,3 @@
-old code
+new code`,
			expected: PatchStats{
				TotalAdditions:      1,
				TotalDeletions:      1,
				MeaningfulAdditions: 1,
				MeaningfulDeletions: 1,
			},
		},
		{
			name: "comment only changes",
			patch: `@@ -1,3 +1,5 @@
 func main() {
+// This is a comment
+// Another comment
 }`,
			expected: PatchStats{
				TotalAdditions:   2,
				CommentAdditions: 2,
			},
		},
		{
			name: "whitespace only changes",
			patch: `@@ -1,3 +1,5 @@
 func main() {
+
+
 }`,
			expected: PatchStats{
				TotalAdditions:      2,
				WhitespaceAdditions: 2,
			},
		},
		{
			name: "mixed meaningful and non-meaningful",
			patch: `@@ -1,5 +1,10 @@
 func main() {
+// Add logging
+    x := 5
+
+    // Calculate result
+    result := x * 2
+
 }`,
			expected: PatchStats{
				TotalAdditions:      6,
				MeaningfulAdditions: 2, // x := 5 and result := x * 2
				CommentAdditions:    2, // two comments
				WhitespaceAdditions: 2, // two empty lines
			},
		},
		{
			name: "deleted comments",
			patch: `@@ -1,5 +1,2 @@
 func main() {
-// Old comment
-/* Block comment */
 }`,
			expected: PatchStats{
				TotalDeletions:   2,
				CommentDeletions: 2,
			},
		},
		{
			name: "python style comments",
			patch: `@@ -1,3 +1,6 @@
 def main():
+# This is a python comment
+"""This is a docstring"""
+    x = 5`,
			expected: PatchStats{
				TotalAdditions:      3,
				MeaningfulAdditions: 1, // x = 5
				CommentAdditions:    2, // # comment and docstring
			},
		},
		{
			name: "sql comments",
			patch: `@@ -1,2 +1,4 @@
 SELECT * FROM users
+-- This is a SQL comment
+WHERE id = 1`,
			expected: PatchStats{
				TotalAdditions:      2,
				MeaningfulAdditions: 1, // WHERE clause
				CommentAdditions:    1, // SQL comment
			},
		},
		{
			name:  "empty patch",
			patch: "",
			expected: PatchStats{
				TotalAdditions:      0,
				TotalDeletions:      0,
				MeaningfulAdditions: 0,
				MeaningfulDeletions: 0,
			},
		},
		{
			name: "context only patch",
			patch: `@@ -1,3 +1,3 @@
 line 1
 line 2
 line 3`,
			expected: PatchStats{
				TotalAdditions:      0,
				TotalDeletions:      0,
				MeaningfulAdditions: 0,
				MeaningfulDeletions: 0,
			},
		},
		{
			name: "header lines should be ignored",
			patch: `--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 context
+new line`,
			expected: PatchStats{
				TotalAdditions:      1,
				MeaningfulAdditions: 1,
			},
		},
		{
			name: "c-style block comment continuation",
			patch: `@@ -1,2 +1,5 @@
 code
+/*
+ * Block comment
+ */`,
			expected: PatchStats{
				TotalAdditions:   3,
				CommentAdditions: 3,
			},
		},
		{
			name: "html comments",
			patch: `@@ -1,2 +1,4 @@
 <div>
+<!-- This is an HTML comment -->
+<p>Content</p>`,
			expected: PatchStats{
				TotalAdditions:      2,
				MeaningfulAdditions: 1, // <p> tag
				CommentAdditions:    1, // HTML comment
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnalyzePatch(tt.patch)
			assert.Equal(t, tt.expected.TotalAdditions, result.TotalAdditions, "TotalAdditions")
			assert.Equal(t, tt.expected.TotalDeletions, result.TotalDeletions, "TotalDeletions")
			assert.Equal(t, tt.expected.MeaningfulAdditions, result.MeaningfulAdditions, "MeaningfulAdditions")
			assert.Equal(t, tt.expected.MeaningfulDeletions, result.MeaningfulDeletions, "MeaningfulDeletions")
			assert.Equal(t, tt.expected.CommentAdditions, result.CommentAdditions, "CommentAdditions")
			assert.Equal(t, tt.expected.CommentDeletions, result.CommentDeletions, "CommentDeletions")
			assert.Equal(t, tt.expected.WhitespaceAdditions, result.WhitespaceAdditions, "WhitespaceAdditions")
			assert.Equal(t, tt.expected.WhitespaceDeletions, result.WhitespaceDeletions, "WhitespaceDeletions")
		})
	}
}

func TestAnalyzePatchSimple(t *testing.T) {
	patch := `@@ -1,3 +1,6 @@
 func main() {
+// comment
+    x := 5
+
+    y := 10
 }`

	adds, dels := AnalyzePatchSimple(patch)
	assert.Equal(t, 2, adds, "meaningful additions (x := 5 and y := 10)")
	assert.Equal(t, 0, dels, "meaningful deletions")
}

func TestIsMeaningfulLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"code line", "x := 5", true},
		{"function definition", "func main() {", true},
		{"return statement", "return nil", true},
		{"comment line", "// comment", false},
		{"empty line", "", false},
		{"whitespace line", "   ", false},
		{"python comment", "# comment", false},
		{"code with leading whitespace", "    x := 5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMeaningfulLine(tt.line)
			assert.Equal(t, tt.expected, result, "IsMeaningfulLine(%q)", tt.line)
		})
	}
}

func TestAnalyzePatch_RealWorldExample(t *testing.T) {
	// Simulate a real-world Go file change
	patch := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -10,6 +10,15 @@ package main
 import "fmt"

+// ProcessData handles data processing
+// It takes input and returns processed output
 func ProcessData(input string) string {
+	// Validate input
+	if input == "" {
+		return ""
+	}
+
+	// Transform the data
+	result := strings.ToUpper(input)
-	return input
+	return result
 }`

	stats := AnalyzePatch(patch)

	// Count what's actually in the patch:
	// Additions (lines starting with +, not +++):
	// 1. +// ProcessData handles data processing  -> comment
	// 2. +// It takes input and returns processed output  -> comment
	// 3. +	// Validate input  -> comment
	// 4. +	if input == ""  -> meaningful
	// 5. +		return ""  -> meaningful
	// 6. +	}  -> meaningful
	// 7. +  (empty line)  -> whitespace
	// 8. +	// Transform the data  -> comment
	// 9. +	result := strings.ToUpper(input)  -> meaningful
	// 10. +	return result  -> meaningful
	// Total: 10 additions, 5 meaningful, 4 comments, 1 whitespace

	// Deletions (lines starting with -, not ---):
	// 1. -	return input  -> meaningful
	// Total: 1 deletion, 1 meaningful

	assert.Equal(t, 10, stats.TotalAdditions, "Total additions")
	assert.Equal(t, 1, stats.TotalDeletions, "Total deletions")
	assert.Equal(t, 5, stats.MeaningfulAdditions, "Meaningful additions")
	assert.Equal(t, 1, stats.MeaningfulDeletions, "Meaningful deletions")
	assert.Equal(t, 4, stats.CommentAdditions, "Comment additions")
	assert.Equal(t, 1, stats.WhitespaceAdditions, "Whitespace additions")
}
