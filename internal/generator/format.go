package generator

import "github.com/epilande/codegrab/internal/secrets"

// FileData holds file content for the generated sections
type FileData struct {
	Path     string
	Content  string
	Language string
	Findings []secrets.Finding
}

// TemplateData is injected into the templates
type TemplateData struct {
	Structure string
	Files     []FileData
	FilePaths []string
}

// Format defines the interface for different output formats
type Format interface {
	// Render converts the template data into the specific format
	Render(data TemplateData) (string, int, error)
	// Extension returns the file extension for this format
	Extension() string
	// Name returns the name of the format
	Name() string
}

// ManifestFormat is optionally implemented by formats that can generate manifests.
// Formats implementing this interface provide regex patterns to find file boundaries
// in the rendered output, enabling manifest generation.
type ManifestFormat interface {
	Format
	// FilePattern returns a regex pattern to find file boundaries in the rendered output.
	// The pattern should have a capture group for the file path.
	FilePattern() string
	// StructurePattern returns a regex pattern to find the structure section boundaries.
	// Returns empty string if the format doesn't have a distinct structure section.
	StructurePattern() string
}
