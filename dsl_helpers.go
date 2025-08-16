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

// SetFileBytes creates or replaces a file at a nested relative path with binary content.
func SetFileBytes(root *Folder, relPath string, data []byte) *File {
	dir, base := filepath.Split(filepath.ToSlash(relPath))
	folder := EnsureFolderPath(root, strings.TrimSuffix(dir, "/"))
	return folder.File(base, FileOptions{Content: data})
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

// SetBytes is a Folder method wrapper for SetFileBytes.
func (f *Folder) SetBytes(relPath string, data []byte) *File { return SetFileBytes(f, relPath, data) }

// GetPath retrieves an entry at a nested path.
func (f *Folder) GetPath(relPath string) (FolderEntry, bool) {
	dir, base := filepath.Split(filepath.ToSlash(relPath))
	parent := EnsureFolderPath(f, strings.TrimSuffix(dir, "/"))
	entry, ok := parent._entries[base]
	return entry, ok
}

// InjectChecksumPath sets a checksum for a file or folder at the given path.
func (f *Folder) InjectChecksumPath(relPath, algo string, digest []byte) bool {
	entry, ok := f.GetPath(relPath)
	if !ok { return false }
	switch v := entry.(type) {
	case *File:
		v.SetChecksum(algo, digest)
		return true
	case *Folder:
		v.SetChecksum(algo, digest)
		return true
	default:
		return false
	}
}