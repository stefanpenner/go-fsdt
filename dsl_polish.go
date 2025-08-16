package fsdt

import (
	"errors"
	"path/filepath"
	"strings"
)

// FS is an alias for NewFolderFromStrings.
func FS(files map[string]string) *Folder { return NewFolderFromStrings(files) }

// Set sets or replaces a file at relPath with the provided string content.
func (f *Folder) Set(relPath string, content string) *File { return SetFileString(f, relPath, content) }

// Mk ensures the nested folder path exists and returns it.
func (f *Folder) Mk(relPath string) *Folder { return EnsureFolderPath(f, relPath) }

// Copy returns a deep clone of this folder.
func (f *Folder) Copy() *Folder { return f.Clone().(*Folder) }

// Exists returns whether a path exists under this folder.
func (f *Folder) Exists(relPath string) bool {
	if relPath == "" || relPath == "." { return true }
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	curr := f
	for i, part := range parts {
		if part == "" { continue }
		entry, ok := curr._entries[part]
		if !ok { return false }
		if i == len(parts)-1 { return true }
		sub, ok := entry.(*Folder)
		if !ok { return false }
		curr = sub
	}
	return true
}

// RemovePath removes a path (file/link/folder) under this folder.
func (f *Folder) RemovePath(relPath string) error {
	dir, base := filepath.Split(filepath.ToSlash(relPath))
	parent := navigateToFolder(f, strings.TrimSuffix(dir, "/"))
	if parent == nil { return errors.New("remove: parent not found") }
	_, ok := parent._entries[base]
	if !ok { return errors.New("remove: entry not found") }
	delete(parent._entries, base)
	return nil
}

// Move renames or relocates an entry from src to dst within this folder tree.
func (f *Folder) Move(src, dst string) error {
	sDir, sBase := filepath.Split(filepath.ToSlash(src))
	dDir, dBase := filepath.Split(filepath.ToSlash(dst))
	sParent := navigateToFolder(f, strings.TrimSuffix(sDir, "/"))
	if sParent == nil { return errors.New("move: source parent not found") }
	entry, ok := sParent._entries[sBase]
	if !ok { return errors.New("move: source not found") }
	dParent := EnsureFolderPath(f, strings.TrimSuffix(dDir, "/"))
	dParent._entries[dBase] = entry
	delete(sParent._entries, sBase)
	return nil
}

func navigateToFolder(root *Folder, path string) *Folder {
	if path == "" || path == "." { return root }
	parts := strings.Split(filepath.ToSlash(path), "/")
	curr := root
	for _, part := range parts {
		if part == "" { continue }
		entry, ok := curr._entries[part]
		if !ok { return nil }
		sub, ok := entry.(*Folder)
		if !ok { return nil }
		curr = sub
	}
	return curr
}

// Store helpers
func Sidecar(baseDir, root, algo string) ChecksumStore { return SidecarStore{BaseDir: baseDir, Root: root, Algorithm: algo} }
func XAttr(key string) ChecksumStore { return XAttrStore{Key: key} }
func Store(stores ...ChecksumStore) ChecksumStore {
	switch len(stores) {
	case 0:
		return nil
	case 1:
		return stores[0]
	default:
		return MultiStore{Stores: stores}
	}
}

// XAttr helpers (on-disk) for tests that write trees to disk
func (f *Folder) XAttrWrite(path, key string, digest []byte) error {
	entry, ok := f.GetPath(path)
	if !ok { return errors.New("xattr write: path not found") }
	switch v := entry.(type) {
	case *File:
		if p, has := v.SourcePath(); has { return writeXAttrChecksum(p, key, digest) }
	case *Folder:
		if v.sourcePath != "" { return writeXAttrChecksum(v.sourcePath, key, digest) }
	}
	return errors.New("xattr write: no on-disk source path")
}

func (f *Folder) XAttrRead(path, key string) ([]byte, bool) {
	entry, ok := f.GetPath(path)
	if !ok { return nil, false }
	switch v := entry.(type) {
	case *File:
		if p, has := v.SourcePath(); has {
			if d, ok, _ := readXAttrChecksum(p, key); ok { return d, true }
		}
	case *Folder:
		if v.sourcePath != "" {
			if d, ok, _ := readXAttrChecksum(v.sourcePath, key); ok { return d, true }
		}
	}
	return nil, false
}