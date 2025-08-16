package fsdt

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	op "github.com/stefanpenner/go-fsdt/operation"
)

type Folder struct {
	// FileInfo might be handy
	_entries map[string]FolderEntry
	mode     os.FileMode
}

var DEFAULT_FOLDER_MODE = os.FileMode(os.ModeDir | 0755)

func NewFolder(cb ...func(f *Folder)) *Folder {
	folder := &Folder{
		_entries: map[string]FolderEntry{},
		mode:     DEFAULT_FOLDER_MODE,
	}

	for _, cb := range cb {
		cb(folder)
	}

	return folder
}

func (f *Folder) Entries() []string {
	entries := make([]string, 0, len(f._entries))
	for name := range f._entries {
		entries = append(entries, name)
	}
	// TODO: sort on demand, then cache until a mutation
	// hopefully a lexigraphic sort
	sort.Strings(entries)
	return entries
}

func (f *Folder) Mode() os.FileMode {
	return f.mode
}

func (f *Folder) RemoveOperation(relativePath string, reason op.Reason) op.Operation {
	operations := make([]op.Operation, 0, len(f._entries))
	for _, entryName := range f.Entries() {
		entry := f._entries[entryName]
		operations = append(operations, entry.RemoveOperation(entryName, reason))
	}

	// TODO: reason
	return op.NewRmdir(relativePath, operations...)
}

func (f *Folder) Get(relativePath string) FolderEntry {
	entry, ok := f._entries[relativePath]
	if ok {
		return entry
	} else {
		panic(fmt.Sprintf("Entry: %s not found in: %v", relativePath, f.Entries()))
	}
}

func (f *Folder) Remove(relativePath string) error {
	_, ok := f._entries[relativePath]
	if ok {
		delete(f._entries, relativePath)
		return nil
	} else {
		return fmt.Errorf("Remove Error: %s not found in: %v", relativePath, f.Entries())
	}
}

func (f *Folder) RemoveChildOperation(relativePath string, reason op.Reason) op.Operation {
	return f.Get(relativePath).RemoveOperation(relativePath, reason)
}

func (f *Folder) CreateOperation(relativePath string, reason op.Reason) op.Operation {
	operations := make([]op.Operation, 0, len(f._entries))

	for _, entryName := range f.Entries() {
		entry := f._entries[entryName]
		operations = append(operations, entry.CreateOperation(entryName, reason))
	}

	return op.NewMkdirOperation(relativePath, operations...)
}

func (f *Folder) ChangeOperation(relativePath string, reason op.Reason, operations ...op.Operation) op.Operation {
	return op.NewChangeFolderOperation(relativePath, operations...)
}

func (f *Folder) CreateChildOperation(relativePath string, reason op.Reason) op.Operation {
	return f.Get(relativePath).CreateOperation(relativePath, reason)
}

func (f *Folder) File(name string, content ...FileOptions) *File {
	file := NewFile(content...)
	f._entries[name] = file
	return file
}

func (f *Folder) FileString(name string, content string) *File {
	file := NewFile(FileOptions{
		Content: []byte(content),
		Mode:    DEFAULT_FILE_MODE,
	})
	f._entries[name] = file
	return file
}

func (f *Folder) Folder(name string, cb ...func(*Folder)) *Folder {
	folder := NewFolder()
	f._entries[name] = folder
	for _, cb := range cb {
		cb(folder)
	}
	return folder
}

func (f *Folder) Symlink(link string, target string) *Link {
	symlink := NewLink(target, SYMLINK)
	f._entries[link] = symlink
	return symlink
}

func (f *Folder) Hardlink(link string, target string) *Link {
	panic("go-fsdt/hardlink unsupported")
}

func (f *Folder) Clone() FolderEntry {
	clone := NewFolder()
	clone.mode = f.mode
	for name, entry := range f._entries {
		clone._entries[name] = entry.Clone()
	}
	return clone
}

func (f *Folder) FileStrings(prefix string) []string {
	entries := []string{}

	for _, name := range f.Entries() {
		entry := f._entries[name]
		fullpath := filepath.Join(prefix, name)

		switch e := entry.(type) {
		case *File, *Link:
			entries = append(entries, fullpath)
		case *Folder:
			entries = append(entries, e.FileStrings(fullpath)...)
		default:
			panic(fmt.Sprintf("go-fsdt/Folder.FileStrings does not support FileEntries of type: %s", e))
		}
	}
	return entries
}

func (f *Folder) Strings(prefix string) []string {
	entries := []string{}

	if prefix != "" {
		// TODO: decide if a non-prefix empty root folder should show up as "/" or just be dropped
		entries = append(entries, prefix+string(os.PathSeparator))
	}

	for _, name := range f.Entries() {
		entry := f._entries[name]
		fullpath := filepath.Join(prefix, name)
		entries = append(entries, entry.Strings(fullpath)...)
	}
	return entries
}

func (f *Folder) WriteTo(location string) error {
	err := os.Mkdir(location, f.mode.Perm())
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	for _, relativePath := range f.Entries() {
		err := f.Get(relativePath).WriteTo(filepath.Join(location, relativePath))
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadFrom(path string) (*Folder, error) {
	folder := NewFolder()
	error := folder.ReadFrom(path)
	return folder, error
}

func (f *Folder) ReadFrom(path string) error {
	dirs, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range dirs {
		if entry.IsDir() {
			folder := f.Folder(entry.Name(), func(f *Folder) {})
			err = folder.ReadFrom(filepath.Join(path, entry.Name()))
			if err != nil {
				return err
			}
		} else if entry.Type().IsRegular() {
			content, err := os.ReadFile(filepath.Join(path, entry.Name()))
			if err != nil {
				return err
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			f.File(entry.Name(), FileOptions{
				Content: content,
				Mode:    info.Mode(),
			})
		} else if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(filepath.Join(path, entry.Name()))
			if err != nil {
				return err
			}

			f.Symlink(entry.Name(), target)
		} else {
			return fmt.Errorf("Unexpected DirEntry Type: %s", entry.Type())
		}
	}

	return nil
}

func (f *Folder) Type() FolderEntryType {
	return FOLDER
}

func (f *Folder) Equal(entry FolderEntry) bool {
	equal, _ := f.EqualWithReason(entry)
	return equal
}

func (f *Folder) EqualWithReason(entry FolderEntry) (bool, op.Reason) {
	// Check if the other entry is also a folder
	otherFolder, isFolder := entry.(*Folder)
	if !isFolder {
		return false, op.Reason{
			Type:   op.TypeChanged,
			Before: f.Type(),
			After:  entry.Type(),
		}
	}

	// Check if modes are different
	if f.mode != otherFolder.mode {
		return false, op.Reason{
			Type:   op.ModeChanged,
			Before: f.mode,
			After:  otherFolder.mode,
		}
	}

	// Check if they have the same number of entries
	if len(f._entries) != len(otherFolder._entries) {
		return false, op.Reason{
			Type:   op.ContentChanged,
			Before: len(f._entries),
			After:  len(otherFolder._entries),
		}
	}

	// Check if all entries are equal
	for name, fEntry := range f._entries {
		otherEntry, exists := otherFolder._entries[name]
		if !exists {
			return false, op.Reason{
				Type:   op.Missing,
				Before: name,
				After:  nil,
			}
		}

		equal, reason := fEntry.EqualWithReason(otherEntry)
		if !equal {
			// Update the path to include the entry name
			if reason.Type != "" {
				reason.Before = name + "/" + fmt.Sprintf("%v", reason.Before)
				reason.After = name + "/" + fmt.Sprintf("%v", reason.After)
			}
			return false, reason
		}
	}

	return true, op.Reason{}
}

func (f *Folder) HasContent() bool {
	return false
}

func (f *Folder) Content() []byte {
	return []byte{}
}

// Checksum is not meaningful for folders; return no checksum.
func (f *Folder) Checksum() ([]byte, string, bool) {
	return nil, "", false
}

func (f *Folder) Diff(b *Folder) op.Operation {
	return Diff(f, b, true)
}

func (f *Folder) CaseInsensitiveDiff(b *Folder) op.Operation {
	return Diff(f, b, false)
}

// DiffFast skips content comparison and only compares structure, type, and mode.
func (f *Folder) DiffFast(b *Folder) op.Operation {
	return DiffWithOptions(f, b, DiffOptions{CaseSensitive: true, ContentStrategy: SkipContent})
}

// DiffPreferChecksums uses checksums (e.g., xattr-provided) when available, falling back to bytes.
func (f *Folder) DiffPreferChecksums(b *Folder, algo string) op.Operation {
	return DiffWithOptions(f, b, DiffOptions{CaseSensitive: true, ContentStrategy: PreferChecksumOrBytes, ChecksumAlgorithm: algo})
}

// DiffWithOptions provides object method access to the configurable diff.
func (f *Folder) DiffWithOptions(b *Folder, opts DiffOptions) op.Operation {
	return DiffWithOptions(f, b, opts)
}

func (f *Folder) ContentString() string {
	panic("fsdt:folder.go(Folder does not implement contentString)")
}
