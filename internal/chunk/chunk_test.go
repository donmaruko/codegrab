package chunk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/epilande/codegrab/internal/generator"
)

type mockFormat struct{}

func (f *mockFormat) Render(data generator.TemplateData) (string, int, error) {
	content := "# Files\n"
	for _, file := range data.Files {
		content += "## " + file.Path + "\n" + file.Content + "\n"
	}
	return content, len(content) / 4, nil
}

func (f *mockFormat) Extension() string { return ".md" }
func (f *mockFormat) Name() string      { return "markdown" }

func TestGroupFilesByDepth(t *testing.T) {
	chunker := NewChunker(&mockFormat{}, 1, "/tmp", "/root")

	files := []generator.FileData{
		{Path: "cmd/main.go", Content: "package main"},
		{Path: "cmd/cli.go", Content: "package main"},
		{Path: "internal/cache/cache.go", Content: "package cache"},
		{Path: "internal/utils/util.go", Content: "package utils"},
		{Path: "README.md", Content: "# README"},
	}

	groups := chunker.GroupFilesByDepth(files)

	if len(groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(groups))
	}

	if len(groups["cmd"]) != 2 {
		t.Errorf("expected 2 files in cmd, got %d", len(groups["cmd"]))
	}

	if len(groups["internal"]) != 2 {
		t.Errorf("expected 2 files in internal, got %d", len(groups["internal"]))
	}

	if len(groups["_root"]) != 1 {
		t.Errorf("expected 1 file in _root, got %d", len(groups["_root"]))
	}
}

func TestGroupFilesByDepth2(t *testing.T) {
	chunker := NewChunker(&mockFormat{}, 2, "/tmp", "/root")

	files := []generator.FileData{
		{Path: "internal/cache/cache.go", Content: "package cache"},
		{Path: "internal/cache/manager.go", Content: "package cache"},
		{Path: "internal/utils/util.go", Content: "package utils"},
	}

	groups := chunker.GroupFilesByDepth(files)

	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}

	if len(groups["internal/cache"]) != 2 {
		t.Errorf("expected 2 files in internal/cache, got %d", len(groups["internal/cache"]))
	}

	if len(groups["internal/utils"]) != 1 {
		t.Errorf("expected 1 file in internal/utils, got %d", len(groups["internal/utils"]))
	}
}

func TestSanitizeFilename(t *testing.T) {
	chunker := NewChunker(&mockFormat{}, 1, "/tmp", "/root")

	tests := []struct {
		input    string
		expected string
	}{
		{"internal/cache", "internal-cache"},
		{"cmd", "cmd"},
		{"_root", "root"},
		{"path/with spaces", "path-with-spaces"},
	}

	for _, tc := range tests {
		result := chunker.sanitizeFilename(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestGenerate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "chunk-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	chunker := NewChunker(&mockFormat{}, 1, tmpDir, "/root")

	data := generator.TemplateData{
		Structure: "project/\n  cmd/\n  internal/\n",
		Files: []generator.FileData{
			{Path: "cmd/main.go", Content: "package main", Language: "go"},
			{Path: "internal/cache.go", Content: "package internal", Language: "go"},
		},
		FilePaths: []string{"cmd/main.go", "internal/cache.go"},
	}

	chunks, err := chunker.Generate(data)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(chunks) != 2 {
		t.Errorf("expected 2 chunks, got %d", len(chunks))
	}

	// Check files were created
	if _, err := os.Stat(filepath.Join(tmpDir, "chunk-cmd.md")); os.IsNotExist(err) {
		t.Error("chunk-cmd.md was not created")
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "chunk-internal.md")); os.IsNotExist(err) {
		t.Error("chunk-internal.md was not created")
	}
}

func TestGenerateIndex(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "chunk-index-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	chunker := NewChunker(&mockFormat{}, 1, tmpDir, "/root")

	chunks := []ChunkInfo{
		{Name: "cmd", Filename: "chunk-cmd.md", FileCount: 2, TokenCount: 500, Files: []string{"main.go", "cli.go"}},
		{Name: "internal", Filename: "chunk-internal.md", FileCount: 5, TokenCount: 2000, Files: []string{"a.go", "b.go", "c.go", "d.go", "e.go"}},
	}

	err = chunker.GenerateIndex("project/\n  cmd/\n  internal/\n", chunks)
	if err != nil {
		t.Fatalf("GenerateIndex failed: %v", err)
	}

	indexPath := filepath.Join(tmpDir, "INDEX.md")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Error("INDEX.md was not created")
	}

	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read INDEX.md: %v", err)
	}

	// Check index contains expected content
	indexStr := string(content)
	if !contains(indexStr, "chunk-cmd.md") {
		t.Error("INDEX.md should contain chunk-cmd.md")
	}
	if !contains(indexStr, "chunk-internal.md") {
		t.Error("INDEX.md should contain chunk-internal.md")
	}
	if !contains(indexStr, "**Total**") {
		t.Error("INDEX.md should contain Total row")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
