package manifest

import (
	"regexp"
	"strings"
	"time"
)

// FileInfo contains the minimal file information needed for manifest building.
type FileInfo struct {
	Path     string
	Language string
}

// FormatInfo contains the format information needed for manifest building.
type FormatInfo struct {
	Name             string
	FilePattern      string
	StructurePattern string
}

// Builder builds manifests from rendered content.
type Builder struct{}

// NewBuilder creates a new manifest builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// BuildFromContent builds a manifest by parsing the rendered content to find file boundaries.
func (b *Builder) BuildFromContent(content string, formatInfo FormatInfo, files []FileInfo) (*Manifest, error) {
	m := &Manifest{
		Version:   "1.0",
		Format:    formatInfo.Name,
		Generated: time.Now().UTC(),
		Files:     make(map[string]FileEntry),
	}

	// Build file path to language map
	langMap := make(map[string]string)
	for _, f := range files {
		langMap[f.Path] = f.Language
	}

	// Find structure section if format supports it
	if formatInfo.StructurePattern != "" {
		m.Structure = b.findStructureSection(content, formatInfo.StructurePattern)
	}

	// Find file boundaries using the format's pattern
	if formatInfo.FilePattern == "" {
		return m, nil
	}

	re := regexp.MustCompile(formatInfo.FilePattern)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	// Filter matches to only include known files
	var validMatches [][]int
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		pathStart, pathEnd := match[2], match[3]
		filePath := content[pathStart:pathEnd]
		// Only include paths that are in our files list
		if _, exists := langMap[filePath]; exists {
			validMatches = append(validMatches, match)
		}
	}

	for i, match := range validMatches {
		// Extract file path from capture group
		pathStart, pathEnd := match[2], match[3]
		filePath := content[pathStart:pathEnd]

		// Byte start is where this file section begins
		byteStart := match[0]

		// Byte end is either the start of the next file or end of content
		var byteEnd int
		if i+1 < len(validMatches) {
			byteEnd = validMatches[i+1][0]
		} else {
			byteEnd = len(content)
		}

		// Trim trailing whitespace from byte range
		byteEnd = trimTrailingWhitespace(content, byteStart, byteEnd)

		// Calculate line numbers
		lineStart := countLines(content, 0, byteStart) + 1
		lineEnd := countLines(content, 0, byteEnd)

		m.Files[filePath] = FileEntry{
			ByteStart: byteStart,
			ByteEnd:   byteEnd,
			LineStart: lineStart,
			LineEnd:   lineEnd,
			Language:  langMap[filePath],
		}
	}

	return m, nil
}

// findStructureSection finds the structure section boundaries using the given pattern.
func (b *Builder) findStructureSection(content string, pattern string) *Section {
	re := regexp.MustCompile(pattern)
	match := re.FindStringIndex(content)
	if match == nil {
		return nil
	}

	byteStart := match[0]
	byteEnd := match[1]

	lineStart := countLines(content, 0, byteStart) + 1
	lineEnd := countLines(content, 0, byteEnd)

	return &Section{
		ByteStart: byteStart,
		ByteEnd:   byteEnd,
		LineStart: lineStart,
		LineEnd:   lineEnd,
	}
}

// countLines counts the number of newlines in content[start:end].
func countLines(content string, start, end int) int {
	return strings.Count(content[start:end], "\n")
}

// trimTrailingWhitespace returns the end position with trailing whitespace trimmed,
// but keeps at least one newline between files.
func trimTrailingWhitespace(content string, start, end int) int {
	// Trim trailing whitespace except for one newline
	trimmed := strings.TrimRight(content[start:end], " \t\r\n")
	newEnd := start + len(trimmed)

	// Add back one newline if there was one
	if newEnd < end && content[newEnd] == '\n' {
		newEnd++
	}

	return newEnd
}
