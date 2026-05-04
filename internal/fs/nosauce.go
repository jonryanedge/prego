package fs

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const NosauceFile = ".nosauce"

type ignoreRule struct {
	dir      string
	patterns []string
}

type parsedPattern struct {
	raw       string
	base      string
	isDirOnly bool
	isNegated bool
	hasSlash  bool
}

func parseNosauceFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var patterns []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, scanner.Err()
}

func loadNosauceRules(dir string) []string {
	path := filepath.Join(dir, NosauceFile)
	patterns, err := parseNosauceFile(path)
	if err != nil {
		return nil
	}
	return patterns
}

func parsePattern(raw string) parsedPattern {
	p := parsedPattern{raw: raw}

	s := raw
	if strings.HasPrefix(s, "!") {
		p.isNegated = true
		s = s[1:]
	}

	if strings.HasSuffix(s, "/") {
		p.isDirOnly = true
		s = strings.TrimSuffix(s, "/")
	}

	p.hasSlash = strings.Contains(s, "/")

	if p.hasSlash {
		p.base = s
	} else {
		p.base = s
	}

	return p
}

func matchIgnoreRule(absPath string, isDir bool, rules []ignoreRule) (bool, string) {
	name := filepath.Base(absPath)

	for _, rule := range rules {
		rel, err := filepath.Rel(rule.dir, absPath)
		if err != nil {
			continue
		}
		if strings.HasPrefix(rel, "..") {
			continue
		}
		if rel == "." {
			continue
		}

		for _, raw := range rule.patterns {
			pp := parsePattern(raw)

			matched := matchPattern(absPath, name, rel, isDir, pp)
			if matched {
				return true, raw
			}
		}
	}

	return false, ""
}

func matchPattern(absPath string, name string, rel string, isDir bool, pp parsedPattern) bool {
	if pp.isDirOnly && !isDir {
		return false
	}

	pattern := pp.base

	if pp.hasSlash {
		return matchSlashedPattern(rel, pattern)
	}

	return matchNamePattern(name, pattern)
}

func matchNamePattern(name string, pattern string) bool {
	if name == pattern {
		return true
	}
	matched, _ := filepath.Match(pattern, name)
	return matched
}

func matchSlashedPattern(rel string, pattern string) bool {
	if rel == pattern {
		return true
	}

	relSlash := filepath.ToSlash(rel)
	patternSlash := filepath.ToSlash(pattern)

	if strings.HasPrefix(patternSlash, "**/") {
		suffix := patternSlash[3:]
		return strings.HasSuffix(relSlash, suffix) || relSlash == suffix
	}

	matched, _ := filepath.Match(pattern, rel)
	return matched
}

func shouldIgnore(absPath string, isDir bool, rules []ignoreRule) bool {
	matched, _ := matchIgnoreRule(absPath, isDir, rules)
	return matched
}
