package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanJekyllContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove jekyll tags",
			input:    "Some text {:target=\"_blank\"} more text",
			expected: "Some text  more text",
		},
		{
			name:     "remove TOC marker",
			input:    "# Title\n- TOC\n{:toc}\nContent",
			expected: "# Title\n\n\nContent",
		},
		{
			name:     "no jekyll content",
			input:    "Plain markdown content",
			expected: "Plain markdown content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(cleanJekyllContent([]byte(tt.input)))
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestExtractImages(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single image",
			content:  "Text ![alt text](/path/to/image.png) more text",
			expected: []string{"/path/to/image.png"},
		},
		{
			name:     "multiple images",
			content:  "![img1](/img1.jpg) and ![img2](/img2.png)",
			expected: []string{"/img1.jpg", "/img2.png"},
		},
		{
			name:     "no images",
			content:  "Just plain text without images",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractImages([]byte(tt.content))

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d images, got %d", len(tt.expected), len(result))
				return
			}

			for i, img := range result {
				if img != tt.expected[i] {
					t.Errorf("expected image '%s', got '%s'", tt.expected[i], img)
				}
			}
		})
	}
}

func TestGetChapters(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a test markdown file with frontmatter
	testContent := `---
title: Test Chapter
nav_order: 1
---

# Test Chapter

This is a test chapter content.

![test image](/images/test.png)
`

	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create markdown converter
	cv := NewMarkdownConverter()

	// Get chapters
	chapters, err := GetChapters(&cv, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify results
	if len(chapters) != 1 {
		t.Errorf("expected 1 chapter, got %d", len(chapters))
	}

	if len(chapters) > 0 {
		ch := chapters[0]
		if ch.Filename != "test" {
			t.Errorf("expected filename 'test', got '%s'", ch.Filename)
		}
		if ch.Meta.Title != "Test Chapter" {
			t.Errorf("expected title 'Test Chapter', got '%s'", ch.Meta.Title)
		}
		if ch.Meta.Order != 1 {
			t.Errorf("expected order 1, got %d", ch.Meta.Order)
		}
		if len(ch.Images) != 1 {
			t.Errorf("expected 1 image, got %d", len(ch.Images))
		}
		if len(ch.Html) == 0 {
			t.Error("expected HTML content but got empty string")
		}
	}
}

func TestConvertMarkdownToHTML(t *testing.T) {
	cv := NewMarkdownConverter()

	content := []byte(`---
title: Test Title
nav_order: 5
---

# Heading

Some **bold** text.
`)

	html, meta, err := convertMarkdownToHTML(&cv, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got '%s'", meta.Title)
	}
	if meta.Order != 5 {
		t.Errorf("expected order 5, got %d", meta.Order)
	}
	if len(html) == 0 {
		t.Error("expected HTML output but got empty string")
	}
}
