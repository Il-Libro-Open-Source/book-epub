package main

import (
	"strings"
	"testing"
)

func TestNewMarkdownConverter(t *testing.T) {
	cv := NewMarkdownConverter()

	if cv == nil {
		t.Fatal("expected markdown converter but got nil")
	}

	// Test basic markdown conversion
	input := []byte("# Test Heading\n\nSome **bold** text.")
	var output strings.Builder

	err := cv.Convert(input, &output)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	html := output.String()
	if !strings.Contains(html, "<h1>") {
		t.Error("expected HTML to contain h1 tag")
	}
	if !strings.Contains(html, "<strong>") {
		t.Error("expected HTML to contain strong tag")
	}
}

func TestMarkdownConverterWithTable(t *testing.T) {
	cv := NewMarkdownConverter()

	input := []byte(`
| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
`)

	var output strings.Builder
	err := cv.Convert(input, &output)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	html := output.String()
	if !strings.Contains(html, "<table>") {
		t.Error("expected HTML to contain table tag")
	}
}

func TestMarkdownConverterWithCodeBlock(t *testing.T) {
	cv := NewMarkdownConverter()

	input := []byte("```go\nfunc main() {\n}\n```")

	var output strings.Builder
	err := cv.Convert(input, &output)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	html := output.String()
	// Il syntax highlighting genera HTML complesso, verifichiamo che ci sia output
	if len(html) == 0 {
		t.Error("expected HTML output for code block")
	}
}

func TestMarkdownConverterXHTML(t *testing.T) {
	cv := NewMarkdownConverter()

	input := []byte("Line 1\n\nLine 2")

	var output strings.Builder
	err := cv.Convert(input, &output)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	html := output.String()
	// Verifichiamo che la conversione abbia generato HTML
	if len(html) == 0 {
		t.Error("expected HTML output")
	}
	if !strings.Contains(html, "<p>") {
		t.Error("expected HTML to contain paragraph tags")
	}
}
