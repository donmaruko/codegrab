package formats

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/epilande/codegrab/internal/generator"
	"github.com/epilande/codegrab/internal/utils"
)

// TxtFormat implements the generator.Format interface for plain text output
type TxtFormat struct{}

// Render converts the template data into plain text format
func (f *TxtFormat) Render(data generator.TemplateData) (string, int, error) {
	tmpl := template.New("txt").Funcs(template.FuncMap{
		"separator": func(s ...string) string {
			length := 60
			if len(s) > 0 && len(s[0]) > length {
				length = len("FILE: " + s[0])
			}
			return strings.Repeat("=", length)
		},
	})

	// Parse the template
	tmpl, err := tmpl.Parse(txtTemplate)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", 0, fmt.Errorf("failed to execute template: %w", err)
	}

	txtContent := buf.String()
	tokenCount := utils.EstimateTokens(txtContent)
	return txtContent, tokenCount, nil
}

// Extension returns the file extension for plain text
func (f *TxtFormat) Extension() string {
	return ".txt"
}

// Name returns the name of the format
func (f *TxtFormat) Name() string {
	return "text"
}

// FilePattern returns a regex pattern to find file boundaries in text output.
// Matches: FILE: path/to/file.go
func (f *TxtFormat) FilePattern() string {
	return `(?m)^FILE: (.+)$`
}

// StructurePattern returns a regex pattern to find the structure section in text output.
// Matches the PROJECT STRUCTURE section.
func (f *TxtFormat) StructurePattern() string {
	return `(?s)^=+\nPROJECT STRUCTURE\n=+\n.*?(?:=+\nPROJECT FILES|$)`
}

// The base template for our generated plain text
const txtTemplate = `{{separator}}
PROJECT STRUCTURE
{{separator}}

{{.Structure}}
{{if .Files}}
{{separator}}
PROJECT FILES
{{separator}}
{{range .Files}}
{{separator .Path}}
FILE: {{.Path}}
{{separator .Path}}

{{.Content}}
{{end}}{{end}}`
