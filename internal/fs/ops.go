package fs

import (
	"fmt"
	"os"
)

func MkdirAll(path string, mode uint32) error {
	perm := os.FileMode(mode)
	if perm == 0 {
		perm = os.FileMode(0755)
	}

	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path %s exists and is not a directory", path)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("checking path %s: %w", path, err)
	}

	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("creating directory %s: %w", path, err)
	}

	return os.Chmod(path, perm)
}

func Symlink(from, to string) error {
	info, err := os.Lstat(to)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			existing, _ := os.Readlink(to)
			if existing == from {
				return nil
			}
			return fmt.Errorf("symlink %s already exists pointing to %s (expected %s)", to, existing, from)
		}
		return fmt.Errorf("path %s already exists and is not a symlink", to)
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("checking symlink %s: %w", to, err)
	}

	parent := ""
	for i := len(to) - 1; i >= 0; i-- {
		if to[i] == '/' {
			parent = to[:i]
			break
		}
	}
	if parent != "" {
		if err := os.MkdirAll(parent, 0755); err != nil {
			return fmt.Errorf("creating parent directory for symlink %s: %w", to, err)
		}
	}

	return os.Symlink(from, to)
}
