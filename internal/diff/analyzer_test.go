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
		// Empty and whitespace - NOT comments (use IsWhitespaceLine instead)
		{"empty string", "", false},
		{"whitespace only", "   ", false},
		{"tab only", "\t", false},
		{"mixed whitespace", "  \t  ", false},

		// C-style comments (Go, Java, JS, C++, etc.)
		{"C single line comment", "// this is a comment", true},
		{"C single line with leading space", "  // this is a comment", true},
		{"C block start", "/* block comment", true},
		{"C block end", "*/", true},
		{"C block continuation", "* continuation", true},
		{"C block continuation with space", "  * continuation", true},
		{"just asterisk", "*", true},
		{"asterisk with slash", "*/", true},

		// Pointer dereferences - NOT comments
		{"pointer dereference", "*ptr = value", false},
		{"pointer in expression", "*foo.bar", false},
		{"multiplication", "*result", false},

		// Doc comments
		{"Rust doc comment", "/// This documents the function", true},
		{"Rust inner doc", "//! Module documentation", true},
		{"JSDoc start", "/** @param x the value */", true},

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

		// Indented code (common in diffs) - NOT comments
		{"tab indented code", "\tfunc main() {", false},
		{"space indented code", "    if x > 0 {", false},
		{"deeply indented", "\t\t\t\treturn nil", false},
		{"mixed indentation", "  \t  for i := range items {", false},
		{"indented closing brace", "\t}", false},
		{"indented method call", "        obj.Method()", false},

		// TypeScript/JavaScript specific - NOT comments
		{"TS interface", "interface User {", false},
		{"TS type alias", "type Handler = () => void;", false},
		{"TS arrow function", "const fn = () => {", false},
		{"TS arrow function with type", "const fn = (x: number): string => {", false},
		{"JS const", "const x = 5;", false},
		{"JS let", "let counter = 0;", false},
		{"JS async", "async function fetch() {", false},
		{"JS await", "const result = await fetch(url);", false},
		{"JS template literal", "const msg = `Hello ${name}`;", false},
		{"JS export", "export default Component;", false},
		{"JS import", "import { useState } from 'react';", false},
		{"TS generic", "function identity<T>(arg: T): T {", false},
		{"React JSX", "<Component prop={value} />", false},
		{"JSX with children", "<div className=\"container\">", false},

		// TypeScript/JavaScript comments
		{"TS comment", "// TypeScript comment", true},
		{"JSDoc block", "/** @type {string} */", true},
		{"TSDoc", "/** @param name - the user name */", true},

		// Go specific - NOT comments
		{"Go struct", "type User struct {", false},
		{"Go interface def", "type Reader interface {", false},
		{"Go func with receiver", "func (u *User) Name() string {", false},
		{"Go goroutine", "go processItem(item)", false},
		{"Go defer", "defer file.Close()", false},
		{"Go channel send", "ch <- value", false},
		{"Go channel receive", "value := <-ch", false},
		{"Go select", "select {", false},
		{"Go case", "case <-done:", false},
		{"Go map literal", "m := map[string]int{}", false},
		{"Go slice literal", "s := []int{1, 2, 3}", false},
		{"Go error handling", "if err != nil {", false},
		{"Go short var decl", "x := 5", false},
		{"Go range", "for i, v := range items {", false},

		// Python specific - NOT comments
		{"Python def", "def main():", false},
		{"Python class", "class User:", false},
		{"Python async def", "async def fetch():", false},
		{"Python decorator", "@property", false},
		{"Python with", "with open('file') as f:", false},
		{"Python try", "try:", false},
		{"Python except", "except ValueError as e:", false},
		{"Python lambda", "fn = lambda x: x * 2", false},
		{"Python list comp", "squares = [x**2 for x in range(10)]", false},
		{"Python dict comp", "d = {k: v for k, v in items}", false},
		{"Python f-string", "msg = f\"Hello {name}\"", false},
		{"Python import from", "from typing import List", false},
		{"Python type hint", "def greet(name: str) -> str:", false},

		// Python comments
		{"Python comment with hash", "# This is a comment", true},
		{"Python inline comment would be code", "x = 5  # inline", false}, // The line starts with code
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

		// Indented code is still meaningful
		{"tab indented code", "\tfunc main() {", true},
		{"deeply indented code", "\t\t\treturn result", true},
		{"space indented code", "        if err != nil {", true},
		{"mixed indentation code", "  \t  for _, item := range items {", true},
		{"indented closing brace", "\t\t}", true},

		// Indented comments are still comments (not meaningful)
		{"indented comment", "\t// TODO: fix this", false},
		{"space indented comment", "    # Python comment", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMeaningfulLine(tt.line)
			assert.Equal(t, tt.expected, result, "IsMeaningfulLine(%q)", tt.line)
		})
	}
}

func TestIsRenameOrMove(t *testing.T) {
	tests := []struct {
		name     string
		fromName string
		toName   string
		expected bool
	}{
		// Rename/move operations - should return true
		{"simple rename", "old.go", "new.go", true},
		{"move to subdirectory", "file.go", "pkg/file.go", true},
		{"move from subdirectory", "pkg/file.go", "file.go", true},
		{"rename in subdirectory", "pkg/old.go", "pkg/new.go", true},
		{"move between directories", "src/file.go", "lib/file.go", true},
		{"complex path rename", "internal/api/v1/handler.go", "internal/api/v2/handler.go", true},

		// NOT rename/move - should return false
		{"new file", "", "new.go", false},
		{"deleted file", "old.go", "", false},
		{"modify same file", "file.go", "file.go", false},
		{"both empty", "", "", false},
		{"same path different case is not rename", "File.go", "File.go", false},

		// Edge cases
		{"whitespace in path rename", "my file.go", "my-file.go", true},
		{"deeply nested rename", "a/b/c/d/e/f.go", "a/b/c/d/e/g.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRenameOrMove(tt.fromName, tt.toName)
			assert.Equal(t, tt.expected, result, "IsRenameOrMove(%q, %q)", tt.fromName, tt.toName)
		})
	}
}

func TestIsDocCommentLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		// Documentation comments
		{"Rust doc comment", "/// This documents the function", true},
		{"Rust doc with leading space", "  /// This documents the function", true},
		{"Rust inner doc", "//! Module documentation", true},
		{"JSDoc block start", "/** @param x the value */", true},
		{"JSDoc block start with space", "  /** @param x */", true},
		{"Python docstring double", "\"\"\"This is a docstring", true},
		{"Python docstring single", "'''This is a docstring", true},
		{"JSDoc annotation line", "* @param x the value", true},
		{"JSDoc annotation with extra space", "*  @returns the result", true},
		{"annotation comment", "// @deprecated use newFunc instead", true},
		{"Python annotation", "# @param x the value", true},

		// Regular comments - NOT doc comments
		{"regular C comment", "// this is a comment", false},
		{"regular Python comment", "# just a comment", false},
		{"block comment start", "/* start of block */", false},
		{"block continuation", "* continuation without annotation", false},

		// Empty and whitespace
		{"empty string", "", false},
		{"whitespace only", "   ", false},

		// Code - NOT doc comments
		{"Go code", "func main() {", false},
		{"Python code", "def main():", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDocCommentLine(tt.line)
			assert.Equal(t, tt.expected, result, "IsDocCommentLine(%q)", tt.line)
		})
	}
}

func TestIsCommentedOutCode(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		// Commented-out code - should return true
		{"commented variable declaration", "// const x = 5;", true},
		{"commented function call", "// fmt.Println(x)", true}, // Ends with )
		{"commented function def", "// func main() {", true},
		{"commented return", "// return nil", true},
		{"commented import", "// import fmt", true},
		{"commented if statement", "// if x > 0 {", true},
		{"commented else", "// else {", true},
		{"commented for loop", "// for i := 0; i < 10; i++ {", true},
		{"commented assignment", "// x = 10", true},       // Contains = operator
		{"commented with equals", "// x = y + 10;", true}, // Ends with ;
		{"Python commented code", "# def main():", true},  // colon at end
		{"commented arrow function", "// const fn = () => {", true},
		{"commented Go assignment", "// x := 5", true},

		// Regular comments - should return false
		{"todo comment", "// TODO: fix this", false},
		{"note comment", "// Note: this is important", false},
		{"explanation comment", "// This function handles errors", false},
		{"section comment", "// ============", false},
		{"url in comment", "// See https://example.com", false},

		// Empty and edge cases
		{"empty string", "", false},
		{"just comment prefix", "//", false},
		{"whitespace only", "   ", false},

		// Code (not commented) - should return false
		{"actual code", "const x = 5;", false},
		{"actual function", "func main() {", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCommentedOutCode(tt.line)
			assert.Equal(t, tt.expected, result, "IsCommentedOutCode(%q)", tt.line)
		})
	}
}
