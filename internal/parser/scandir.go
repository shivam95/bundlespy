package parser

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// DefaultExtensions is the allowlist of file extensions bundlespy tracks by default.
// Override per-invocation with the --ext flag.
var DefaultExtensions = []string{
	// JavaScript
	".js", ".mjs", ".cjs",
	// CSS
	".css",
	// Images
	".webp", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".avif",
	// Fonts
	".woff", ".woff2", ".ttf", ".eot", ".otf",
}

// ScanDir walks a build output directory and returns the sizes of tracked files.
// Only extensions in exts are included; pass nil to use DefaultExtensions.
// Asset names are stored relative to dir so baselines are portable across machines.
func ScanDir(dir string, exts []string) (*BuildStats, error) {
	if len(exts) == 0 {
		exts = DefaultExtensions
	}

	var assets []Asset

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if !shouldInclude(rel, exts) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		assets = append(assets, Asset{
			Name:    rel,
			Size:    info.Size(),
			IsChunk: isChunk(rel),
		})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", dir, err)
	}
	if len(assets) == 0 {
		return nil, fmt.Errorf("scan %s: no files found", dir)
	}

	return &BuildStats{Tool: "dir", Assets: assets}, nil
}

// shouldInclude returns true if the file's extension is in the allowlist.
func shouldInclude(name string, exts []string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	for _, e := range exts {
		if ext == e {
			return true
		}
	}
	return false
}

// isChunk returns true for JS and CSS files — the assets that drive bundle size.
func isChunk(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".js" || ext == ".css"
}
