package parser

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// ScanDir walks a build output directory and returns the sizes of all files.
// Source maps (.map) are excluded — they are not served to users.
// Asset names are stored relative to dir so baselines are portable across machines.
func ScanDir(dir string) (*BuildStats, error) {
	var assets []Asset

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".map") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(dir, path)
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

// isChunk returns true for JS and CSS files — the assets that drive bundle size.
func isChunk(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".js" || ext == ".css"
}
