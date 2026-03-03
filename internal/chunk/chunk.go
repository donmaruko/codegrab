package chunk

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/epilande/codegrab/internal/generator"
	"github.com/epilande/codegrab/internal/utils"
)

// ChunkInfo holds metadata about a generated chunk
type ChunkInfo struct {
	Name       string
	Filename   string
	Files      []string
	FileCount  int
	TokenCount int
}

// Chunker handles splitting output into multiple files by directory
type Chunker struct {
	format     generator.Format
	depth      int
	outputDir  string
	rootPath   string
}

// NewChunker creates a new Chunker
func NewChunker(format generator.Format, depth int, outputDir string, rootPath string) *Chunker {
	return &Chunker{
		format:    format,
		depth:     depth,
		outputDir: outputDir,
		rootPath:  rootPath,
	}
}

// GroupFilesByDepth groups files by their directory at the specified depth
func (c *Chunker) GroupFilesByDepth(files []generator.FileData) map[string][]generator.FileData {
	groups := make(map[string][]generator.FileData)

	for _, file := range files {
		key := c.getGroupKey(file.Path)
		groups[key] = append(groups[key], file)
	}

	return groups
}

// getGroupKey returns the directory path at the configured depth
func (c *Chunker) getGroupKey(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")

	if len(parts) <= c.depth {
		// File is at or above the chunk depth, use parent dir or "root"
		if len(parts) == 1 {
			return "_root"
		}
		return strings.Join(parts[:len(parts)-1], "/")
	}

	return strings.Join(parts[:c.depth], "/")
}

// Generate creates chunked output files and returns chunk info for the index
func (c *Chunker) Generate(data generator.TemplateData) ([]ChunkInfo, error) {
	// Create output directory
	if err := os.MkdirAll(c.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Group files by depth
	groups := c.GroupFilesByDepth(data.Files)

	// Sort group keys for consistent output
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var chunks []ChunkInfo

	for _, key := range keys {
		files := groups[key]

		// Create chunk data
		chunkData := generator.TemplateData{
			Structure: "", // No structure in individual chunks
			Files:     files,
			FilePaths: make([]string, len(files)),
		}
		for i, f := range files {
			chunkData.FilePaths[i] = f.Path
		}

		// Render the chunk
		content, tokenCount, err := c.format.Render(chunkData)
		if err != nil {
			return nil, fmt.Errorf("failed to render chunk %s: %w", key, err)
		}

		// Generate chunk filename
		chunkName := c.sanitizeFilename(key)
		chunkFilename := fmt.Sprintf("chunk-%s%s", chunkName, c.format.Extension())
		chunkPath := filepath.Join(c.outputDir, chunkFilename)

		// Write chunk file
		if err := os.WriteFile(chunkPath, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write chunk %s: %w", chunkFilename, err)
		}

		// Collect file paths for chunk info
		filePaths := make([]string, len(files))
		for i, f := range files {
			filePaths[i] = f.Path
		}

		chunks = append(chunks, ChunkInfo{
			Name:       key,
			Filename:   chunkFilename,
			Files:      filePaths,
			FileCount:  len(files),
			TokenCount: tokenCount,
		})
	}

	return chunks, nil
}

// GenerateIndex creates the INDEX.md file
func (c *Chunker) GenerateIndex(structure string, chunks []ChunkInfo) error {
	var sb strings.Builder

	sb.WriteString("# Project Index\n\n")
	sb.WriteString("## Structure\n\n")
	sb.WriteString("```\n")
	sb.WriteString(structure)
	sb.WriteString("```\n\n")

	sb.WriteString("## Chunks\n\n")
	sb.WriteString("| Chunk | Files | Tokens | Contents |\n")
	sb.WriteString("|-------|-------|--------|----------|\n")

	totalFiles := 0
	totalTokens := 0

	for _, chunk := range chunks {
		// Truncate contents preview
		contentsPreview := c.getContentsPreview(chunk.Files, 50)
		sb.WriteString(fmt.Sprintf("| [%s](%s) | %d | %s | %s |\n",
			chunk.Filename,
			chunk.Filename,
			chunk.FileCount,
			utils.FormatTokenCount(chunk.TokenCount),
			contentsPreview,
		))
		totalFiles += chunk.FileCount
		totalTokens += chunk.TokenCount
	}

	sb.WriteString(fmt.Sprintf("| **Total** | **%d** | **%s** | |\n",
		totalFiles,
		utils.FormatTokenCount(totalTokens),
	))

	sb.WriteString("\n## Usage\n\n")
	sb.WriteString("1. Find the relevant chunk in the table above\n")
	sb.WriteString("2. Open the chunk file\n")
	sb.WriteString("3. Copy and paste into your LLM\n")

	indexPath := filepath.Join(c.outputDir, "INDEX.md")
	if err := os.WriteFile(indexPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write INDEX.md: %w", err)
	}

	return nil
}

// sanitizeFilename converts a path to a safe filename
func (c *Chunker) sanitizeFilename(path string) string {
	// Replace slashes with dashes
	name := strings.ReplaceAll(path, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")

	// Remove or replace other problematic characters
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, " ", "-")

	// Handle root files
	if name == "_root" {
		return "root"
	}

	return name
}

// getContentsPreview returns a truncated preview of file names
func (c *Chunker) getContentsPreview(files []string, maxLen int) string {
	if len(files) == 0 {
		return ""
	}

	// Get just the filenames
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = filepath.Base(f)
	}

	preview := strings.Join(names, ", ")
	if len(preview) > maxLen {
		preview = preview[:maxLen-3] + "..."
	}

	return preview
}
