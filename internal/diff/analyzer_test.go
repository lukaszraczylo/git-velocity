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
