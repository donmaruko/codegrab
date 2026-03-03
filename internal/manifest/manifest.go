package manifest

import "time"

// Manifest contains metadata about the generated context file,
// including byte/line offsets for each file for efficient random access.
type Manifest struct {
	Version   string               `json:"version"`
	Format    string               `json:"format"`
	Generated time.Time            `json:"generated"`
	Structure *Section             `json:"structure,omitempty"`
	Files     map[string]FileEntry `json:"files"`
}

// FileEntry contains the byte and line offsets for a single file in the output.
type FileEntry struct {
	ByteStart int    `json:"byteStart"`
	ByteEnd   int    `json:"byteEnd"`
	LineStart int    `json:"lineStart"`
	LineEnd   int    `json:"lineEnd"`
	Language  string `json:"language,omitempty"`
}

// Section contains byte and line offsets for a section of the output (e.g., structure).
type Section struct {
	ByteStart int `json:"byteStart"`
	ByteEnd   int `json:"byteEnd"`
	LineStart int `json:"lineStart"`
	LineEnd   int `json:"lineEnd"`
}
