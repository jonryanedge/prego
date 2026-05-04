package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

var ValidCategories = map[string]bool{
	"core":      true,
	"documents": true,
	"repos":     true,
}

func Validate(cfg *Config) error {
	if cfg.Version != Version {
		return fmt.Errorf("unsupported config version: %d (expected %d)", cfg.Version, Version)
	}

	if len(cfg.Directory) == 0 {
		return fmt.Errorf("no directory categories defined")
	}

	seenPaths := make(map[string]string)

	for cat, dirCat := range cfg.Directory {
		if dirCat.Root == "" {
			return fmt.Errorf("category %q: root is required", cat)
		}

		resolvedRoot := ResolveRoot(dirCat.Root)

		for _, entry := range dirCat.Entries {
			if entry.Path == "" {
				return fmt.Errorf("category %q: entry path is empty", cat)
			}

			if !strings.HasPrefix(entry.Path, "/") && !strings.HasPrefix(entry.Path, "~/") {
				if strings.HasPrefix(entry.Path, "..") {
					return fmt.Errorf("category %q: path %q must not escape the root", cat, entry.Path)
				}
			}

			if entry.Mode != 0 && entry.Mode > 0777 {
				return fmt.Errorf("category %q: path %q has invalid mode %04o", cat, entry.Path, entry.Mode)
			}

			expanded := ResolveEntryPath(entry.Path, resolvedRoot)
			norm := filepath.Clean(expanded)
			if prev, exists := seenPaths[norm]; exists {
				return fmt.Errorf("duplicate path %q in categories %q and %q", entry.Path, prev, cat)
			}
			seenPaths[norm] = cat
		}

		for i, sl := range dirCat.Symlinks {
			if sl.From == "" {
				return fmt.Errorf("category %q: symlink[%d].from is empty", cat, i)
			}
			if sl.To == "" {
				return fmt.Errorf("category %q: symlink[%d].to is empty", cat, i)
			}
		}
	}

	for _, cmd := range cfg.System.Hooks.PostCreate {
		if strings.TrimSpace(cmd) == "" {
			return fmt.Errorf("system.hooks.post_create contains empty command")
		}
	}

	return nil
}
