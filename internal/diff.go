package fs

import (
	"fmt"
	"os"

	"github.com/jonryanedge/prego/internal/config"
)

type DriftType string

const (
	MissingDir   DriftType = "missing_dir"
	ExtraDir     DriftType = "extra_dir"
	ModeMismatch DriftType = "mode_mismatch"
	LinkMismatch DriftType = "link_mismatch"
	LinkMissing  DriftType = "link_missing"
)

type Drift struct {
	Type     DriftType
	Category string
	Path     string
	Expected string
	Actual   string
}

func Diff(cfg *config.Config) []Drift {
	var drifts []Drift

	for cat, dirCat := range cfg.Dirs {
		for _, entry := range dirCat.Entries {
			expanded := config.ExpandPath(entry.Path)
			info, err := os.Stat(expanded)
			if err != nil {
				if os.IsNotExist(err) {
					expectedMode := entry.Mode
					if expectedMode == 0 {
						expectedMode = 0755
					}
					drifts = append(drifts, Drift{
						Type:     MissingDir,
						Category: cat,
						Path:     entry.Path,
						Expected: fmt.Sprintf("directory (mode %04o)", expectedMode),
						Actual:   "does not exist",
					})
				}
				continue
			}

			if !info.IsDir() {
				drifts = append(drifts, Drift{
					Type:     ExtraDir,
					Category: cat,
					Path:     entry.Path,
					Expected: "directory",
					Actual:   "file",
				})
				continue
			}

			expectedMode := entry.Mode
			if expectedMode == 0 {
				expectedMode = 0755
			}
			actualMode := uint32(info.Mode().Perm())
			if actualMode != expectedMode {
				drifts = append(drifts, Drift{
					Type:     ModeMismatch,
					Category: cat,
					Path:     entry.Path,
					Expected: fmt.Sprintf("%04o", expectedMode),
					Actual:   fmt.Sprintf("%04o", actualMode),
				})
			}
		}

		for _, sl := range dirCat.Symlinks {
			expandedTo := config.ExpandPath(sl.To)
			info, err := os.Lstat(expandedTo)
			if err != nil {
				if os.IsNotExist(err) {
					drifts = append(drifts, Drift{
						Type:     LinkMissing,
						Category: cat,
						Path:     sl.To,
						Expected: fmt.Sprintf("symlink -> %s", sl.From),
						Actual:   "does not exist",
					})
				}
				continue
			}

			if info.Mode()&os.ModeSymlink == 0 {
				drifts = append(drifts, Drift{
					Type:     LinkMismatch,
					Category: cat,
					Path:     sl.To,
					Expected: fmt.Sprintf("symlink -> %s", sl.From),
					Actual:   "not a symlink",
				})
				continue
			}

			actual, err := os.Readlink(expandedTo)
			if err != nil {
				continue
			}
			expandedFrom := config.ExpandPath(sl.From)
			if actual != expandedFrom {
				drifts = append(drifts, Drift{
					Type:     LinkMismatch,
					Category: cat,
					Path:     sl.To,
					Expected: expandedFrom,
					Actual:   actual,
				})
			}
		}
	}

	return drifts
}
