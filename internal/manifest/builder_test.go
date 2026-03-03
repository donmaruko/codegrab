package manifest

import (
	"testing"
)

func TestBuilderMarkdown(t *testing.T) {
	content := `# Project Structure

` + "```" + `
project/
├── main.go
└── lib/
    └── util.go
` + "```" + `

# Project Files

## File: ` + "`main.go`" + `

` + "```go" + `
package main

func main() {}
` + "```" + `

## File: ` + "`lib/util.go`" + `

` + "```go" + `
package lib

func Helper() {}
` + "```" + `
`

	files := []FileInfo{
		{Path: "main.go", Language: "go"},
		{Path: "lib/util.go", Language: "go"},
	}

	formatInfo := FormatInfo{
		Name:             "markdown",
		FilePattern:      "(?m)^## File: `([^`]+)`",
		StructurePattern: "(?s)^# Project Structure\\n\\n```\\n.*?```",
	}

	builder := NewBuilder()
	m, err := builder.BuildFromContent(content, formatInfo, files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", m.Version)
	}

	if m.Format != "markdown" {
		t.Errorf("expected format markdown, got %s", m.Format)
	}

	if m.Structure == nil {
		t.Fatal("expected structure to be set")
	}

	if m.Structure.LineStart != 1 {
		t.Errorf("expected structure lineStart 1, got %d", m.Structure.LineStart)
	}

	if len(m.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(m.Files))
	}

	mainEntry, ok := m.Files["main.go"]
	if !ok {
		t.Fatal("expected main.go in files")
	}

	if mainEntry.Language != "go" {
		t.Errorf("expected language go, got %s", mainEntry.Language)
	}

	if mainEntry.ByteStart <= 0 {
		t.Errorf("expected positive byteStart, got %d", mainEntry.ByteStart)
	}

	utilEntry, ok := m.Files["lib/util.go"]
	if !ok {
		t.Fatal("expected lib/util.go in files")
	}

	if utilEntry.ByteStart <= mainEntry.ByteEnd {
		t.Errorf("expected util byteStart > main byteEnd")
	}
}

func TestBuilderXML(t *testing.T) {
	content := `<?xml version="1.0" encoding="UTF-8"?>
<project>
  <filesystem>
    <directory name=".">
      <file name="main.go" />
    </directory>
  </filesystem>
  <files>
    <file path="main.go" language="go"><![CDATA[
package main

func main() {}
]]></file>
    <file path="lib/util.go" language="go"><![CDATA[
package lib

func Helper() {}
]]></file>
  </files>
</project>`

	files := []FileInfo{
		{Path: "main.go", Language: "go"},
		{Path: "lib/util.go", Language: "go"},
	}

	formatInfo := FormatInfo{
		Name:             "xml",
		FilePattern:      `<file path="([^"]+)"`,
		StructurePattern: `(?s)<filesystem>.*?</filesystem>`,
	}

	builder := NewBuilder()
	m, err := builder.BuildFromContent(content, formatInfo, files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(m.Files))
	}

	if m.Structure == nil {
		t.Fatal("expected structure to be set")
	}
}

func TestBuilderText(t *testing.T) {
	content := `============================================================
PROJECT STRUCTURE
============================================================

project/
├── main.go
└── lib/
    └── util.go

============================================================
PROJECT FILES
============================================================
============================================================
FILE: main.go
============================================================

package main

func main() {}

============================================================
FILE: lib/util.go
============================================================

package lib

func Helper() {}
`

	files := []FileInfo{
		{Path: "main.go", Language: "go"},
		{Path: "lib/util.go", Language: "go"},
	}

	formatInfo := FormatInfo{
		Name:             "text",
		FilePattern:      `(?m)^FILE: (.+)$`,
		StructurePattern: `(?s)^=+\nPROJECT STRUCTURE\n=+\n.*?(?:=+\nPROJECT FILES|$)`,
	}

	builder := NewBuilder()
	m, err := builder.BuildFromContent(content, formatInfo, files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(m.Files))
	}

	mainEntry := m.Files["main.go"]
	utilEntry := m.Files["lib/util.go"]

	if mainEntry.ByteStart >= utilEntry.ByteStart {
		t.Errorf("expected main.go before lib/util.go")
	}
}

func TestBuilderNoManifestSupport(t *testing.T) {
	content := "some content"
	files := []FileInfo{{Path: "test.go", Language: "go"}}

	// Format with no patterns
	formatInfo := FormatInfo{
		Name:             "unknown",
		FilePattern:      "",
		StructurePattern: "",
	}

	builder := NewBuilder()
	m, err := builder.BuildFromContent(content, formatInfo, files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Format != "unknown" {
		t.Errorf("expected format unknown, got %s", m.Format)
	}

	if len(m.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(m.Files))
	}

	if m.Structure != nil {
		t.Errorf("expected no structure")
	}
}

func TestBuilderFiltersUnknownFiles(t *testing.T) {
	// Content contains a file header that's not in our files list
	content := `## File: ` + "`known.go`" + `

package known

## File: ` + "`unknown.go`" + `

package unknown
`

	files := []FileInfo{
		{Path: "known.go", Language: "go"},
		// Note: unknown.go is NOT in the list
	}

	formatInfo := FormatInfo{
		Name:        "markdown",
		FilePattern: "(?m)^## File: `([^`]+)`",
	}

	builder := NewBuilder()
	m, err := builder.BuildFromContent(content, formatInfo, files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(m.Files))
	}

	if _, ok := m.Files["known.go"]; !ok {
		t.Error("expected known.go in files")
	}

	if _, ok := m.Files["unknown.go"]; ok {
		t.Error("did not expect unknown.go in files")
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		content  string
		start    int
		end      int
		expected int
	}{
		{"a\nb\nc", 0, 5, 2},
		{"no newlines", 0, 11, 0},
		{"a\n\n\nb", 0, 5, 3},
		{"", 0, 0, 0},
	}

	for _, tc := range tests {
		result := countLines(tc.content, tc.start, tc.end)
		if result != tc.expected {
			t.Errorf("countLines(%q, %d, %d) = %d, want %d",
				tc.content, tc.start, tc.end, result, tc.expected)
		}
	}
}
