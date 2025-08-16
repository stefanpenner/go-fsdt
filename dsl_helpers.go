package fsdt

import (
	"path/filepath"
	"strconv"
	"strings"
)

// EnsureFolderPath walks/creates nested folders under root for the given path (e.g., "a/b/c").
func EnsureFolderPath(root *Folder, path string) *Folder {
	current := root
	if path == "" || path == "." { return current }
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, part := range parts {
		if part == "" { continue }
		entry := current._entries[part]
		if sub, ok := entry.(*Folder); ok {
			current = sub
			continue
		}
		sub := NewFolder()
		current._entries[part] = sub
		current = sub
	}
	return current
}

// SetFileString creates or replaces a file at a nested relative path with string content.
func SetFileString(root *Folder, relPath string, content string) *File {
	dir, base := filepath.Split(filepath.ToSlash(relPath))
	folder := EnsureFolderPath(root, strings.TrimSuffix(dir, "/"))
	return folder.FileString(base, content)
}

// NewFolderFromStrings builds a folder from a map of path->string content.
func NewFolderFromStrings(files map[string]string) *Folder {
	root := NewFolder()
	for p, c := range files {
		SetFileString(root, p, c)
	}
	return root
}

// BuildDeepFiles creates a folder tree of depth d and width w with simple contents.
func BuildDeepFiles(depth, width int) *Folder {
	root := NewFolder()
	current := root
	for d := 0; d < depth; d++ {
		name := "d_" + strconv.Itoa(d)
		sub := current.Folder(name)
		for i := 0; i < width; i++ {
			file := "f_" + strconv.Itoa(i) + ".txt"
			sub.FileString(file, "data")
		}
		current = sub
	}
	return root
}

// CopyTreeVirtual returns a deep clone of the folder (virtual copy).
func CopyTreeVirtual(src *Folder) *Folder { return src.Clone().(*Folder) }