package fs

import (
	"os"
	"path/filepath"

	"github.com/jonryanedge/prego/internal/config"
)

type ScanEntry struct {
	Path   string
	Mode   uint32
	VCS    string
	Remote string
}

type IgnoredEntry struct {
	Path    string
	Pattern string
	Source  string
}

type ScanResult struct {
	Entries []ScanEntry
	Ignored []IgnoredEntry
}

func Scan(root string, depth int) (*ScanResult, error) {
	expanded := config.ExpandPath(root)

	abs, err := filepath.Abs(expanded)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return &ScanResult{}, nil
	}

	result := &ScanResult{}
	var rules []ignoreRule

	err = filepath.WalkDir(abs, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return nil
		}

		if rel == "." {
			patterns := loadNosauceRules(path)
			if patterns != nil {
				rules = append(rules, ignoreRule{dir: path, patterns: patterns})
			}
			return nil
		}

		basename := filepath.Base(path)
		if basename == ".git" || basename == NosauceFile {
			if basename == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		if matched, pattern := matchIgnoreRule(path, true, rules); matched {
			result.Ignored = append(result.Ignored, IgnoredEntry{
				Path:    path,
				Pattern: pattern,
				Source:  NosauceFile,
			})
			return filepath.SkipDir
		}

		patterns := loadNosauceRules(path)
		if patterns != nil {
			rules = append(rules, ignoreRule{dir: path, patterns: patterns})
		}

		fi, err := d.Info()
		if err != nil {
			return nil
		}

		entry := ScanEntry{
			Path: path,
			Mode: uint32(fi.Mode().Perm()),
		}

		if IsGitRepo(path) {
			entry.VCS = "git"
			entry.Remote = GitRemoteURL(path)
			result.Entries = append(result.Entries, entry)
			return filepath.SkipDir
		}

		if depth > 0 {
			if pathDepth(rel) > depth {
				return filepath.SkipDir
			}
		}

		result.Entries = append(result.Entries, entry)
		return nil
	})

	return result, err
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
