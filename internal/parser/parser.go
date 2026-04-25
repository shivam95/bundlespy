package parser

// Asset represents a single output file from a build (JS chunk, CSS file, font, etc.)
type Asset struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	IsChunk bool   `json:"is_chunk"`
}

// BuildStats is the normalized result of scanning a build output directory.
type BuildStats struct {
	Tool      string // always "dir"
	BuildTime int64  // always 0 (not available from filesystem)
	Hash      string // always "" (not available from filesystem)
	Assets    []Asset
}
