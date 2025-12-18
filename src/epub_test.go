package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewBookBuilder(t *testing.T) {
	cfg := &Config{
		Input:  "/tmp/test",
		Output: "./test.epub",
		Cover:  "./test-cover.jpg",
		Style:  "./test-style.css",
		Uuid:   "test-uuid-1234",
	}

	builder, err := NewBookBuilder(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if builder == nil {
		t.Fatal("expected builder but got nil")
	}
	if builder.book == nil {
		t.Fatal("expected book but got nil")
	}
	if builder.config != cfg {
		t.Error("config not properly set")
	}
}

func TestSetupMetadata(t *testing.T) {
	cfg := &Config{
		Uuid: "test-uuid-1234",
	}

	builder, err := NewBookBuilder(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = builder.SetupMetadata()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// The metadata is set internally, we just verify no error occurred
}

func TestAddCover(t *testing.T) {
	// Create a temporary test image
	tmpDir := t.TempDir()
	coverPath := filepath.Join(tmpDir, "cover.jpg")

	// Create a minimal valid JPEG file (1x1 pixel)
	jpegData := []byte{
		0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46,
		0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x48,
		0x00, 0x48, 0x00, 0x00, 0xFF, 0xD9,
	}
	if err := os.WriteFile(coverPath, jpegData, 0644); err != nil {
		t.Fatalf("failed to create test cover: %v", err)
	}

	cfg := &Config{
		Cover: coverPath,
		Uuid:  "test-uuid",
	}

	builder, err := NewBookBuilder(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = builder.AddCover()
	if err != nil {
		t.Errorf("unexpected error adding cover: %v", err)
	}
}

func TestAddStylesheet(t *testing.T) {
	// Create a temporary test CSS file
	tmpDir := t.TempDir()
	cssPath := filepath.Join(tmpDir, "style.css")
	cssContent := []byte("body { margin: 0; }")
	if err := os.WriteFile(cssPath, cssContent, 0644); err != nil {
		t.Fatalf("failed to create test CSS: %v", err)
	}

	cfg := &Config{
		Style: cssPath,
		Uuid:  "test-uuid",
	}

	builder, err := NewBookBuilder(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	internalCssPath, err := builder.AddStylesheet()
	if err != nil {
		t.Errorf("unexpected error adding stylesheet: %v", err)
	}
	if internalCssPath == "" {
		t.Error("expected non-empty CSS path")
	}
}

func TestAddImagesEmpty(t *testing.T) {
	cfg := &Config{
		Input: "/tmp/test",
		Uuid:  "test-uuid",
	}

	builder, err := NewBookBuilder(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with empty chapters
	chapters := []*Chapter{}
	err = builder.AddImages(chapters, cfg.Input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddImagesWithChapters(t *testing.T) {
	// Create a temporary test image
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "test.png")

	// Create a minimal valid PNG file
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x44, 0x41,
		0x54, 0x08, 0x99, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
		0x42, 0x60, 0x82,
	}
	if err := os.WriteFile(imagePath, pngData, 0644); err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}

	cfg := &Config{
		Input: tmpDir,
		Uuid:  "test-uuid",
	}

	builder, err := NewBookBuilder(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	chapters := []*Chapter{
		{
			Filename: "test",
			Html:     "<p>Test <img src=\"/book/test.png\" /></p>",
			Images:   []string{"/book/test.png"},
			Meta:     ChapterMeta{Title: "Test", Order: 1},
		},
	}

	err = builder.AddImages(chapters, tmpDir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify that the HTML was updated
	if chapters[0].Html == "<p>Test <img src=\"/book/test.png\" /></p>" {
		t.Error("expected HTML to be updated with internal image path")
	}
}

func TestCreateChaptersEmpty(t *testing.T) {
	cfg := &Config{
		Uuid: "test-uuid",
	}

	builder, err := NewBookBuilder(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	chapters := []*Chapter{}
	err = builder.CreateChapters(chapters, "style.css")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateChaptersWithContent(t *testing.T) {
	cfg := &Config{
		Uuid: "test-uuid",
	}

	builder, err := NewBookBuilder(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	chapters := []*Chapter{
		{
			Filename: "chapter1",
			Html:     "<h1>Chapter 1</h1><p>Content</p>",
			Meta:     ChapterMeta{Title: "Chapter 1", Order: 1},
		},
	}

	err = builder.CreateChapters(chapters, "style.css")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.epub")

	cfg := &Config{
		Output: outputPath,
		Uuid:   "test-uuid",
	}

	builder, err := NewBookBuilder(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = builder.SetupMetadata()
	if err != nil {
		t.Fatalf("unexpected error setting metadata: %v", err)
	}

	err = builder.Save()
	if err != nil {
		t.Errorf("unexpected error saving: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("expected EPUB file to be created")
	}
}
