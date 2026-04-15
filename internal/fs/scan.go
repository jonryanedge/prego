package fs

import (
	"os"
	"path/filepath"

	"github.com/jonryanedge/prego/internal/config"
)

type ScanEntry struct {
	Path string
	Mode uint32
}

func Scan(root string, depth int) ([]ScanEntry, error) {
	expanded := config.ExpandPath(root)

	info, err := os.Stat(expanded)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil
	}

	var entries []ScanEntry
	seen := make(map[string]bool)

	err = filepath.WalkDir(expanded, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(expanded, path)
		if err != nil {
			return nil
		}

		if rel == "." {
			return nil
		}

		if depth > 0 {
			d := pathDepth(rel)
			if d > depth {
				return nil
			}
		}

		abs := filepath.Join(expanded, rel)
		if seen[abs] {
			return nil
		}
		seen[abs] = true

		fi, err := d.Info()
		if err != nil {
			return nil
		}

		entries = append(entries, ScanEntry{
			Path: abs,
			Mode: uint32(fi.Mode().Perm()),
		})
		return nil
	})

	return entries, err
}

func pathDepth(rel string) int {
	if rel == "." {
		return 0
	}
	count := 0
	for _, c := range rel {
		if c == os.PathSeparator {
			count++
		}
	}
	return count + 1
}
