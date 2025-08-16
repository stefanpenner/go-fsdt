package fsdt

import (
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

func normalizePath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return filepath.ToSlash(filepath.Join(prefix, name))
}

func shouldExclude(path string, excludes []string) bool {
	if len(excludes) == 0 {
		return false
	}
	p := filepath.ToSlash(path)
	for _, g := range excludes {
		if ok, _ := doublestar.PathMatch(g, p); ok {
			return true
		}
	}
	return false
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) { return false }
	m := map[string]int{}
	for _, s := range a { m[s]++ }
	for _, s := range b { if m[s] == 0 { return false }; m[s]-- }
	for _, v := range m { if v != 0 { return false } }
	return true
}