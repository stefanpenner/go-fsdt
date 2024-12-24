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
	entries := []string{}
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

func (f *Folder) RemoveOperation(relativePath string) op.Operation {
	operations := []op.Operation{}
	for _, relativePath := range f.Entries() {
		entry := f._entries[relativePath]
		operations = append(operations, entry.RemoveOperation(relativePath))
	}
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

func (f *Folder) RemoveChildOperation(relativePath string) op.Operation {
	return f.Get(relativePath).RemoveOperation(relativePath)
}

func (f *Folder) CreateOperation(relativePath string) op.Operation {
	operations := []op.Operation{}

	for _, relativePath := range f.Entries() {
		entry := f._entries[relativePath]
		operations = append(operations, entry.CreateOperation(relativePath))
	}

	return op.NewMkdirOperation(relativePath, operations...)
}

func (f *Folder) CreateChildOperation(relativePath string) op.Operation {
	return f.Get(relativePath).CreateOperation(relativePath)
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
		case *File:
			entries = append(entries, e.Strings(fullpath)...)
		case *Folder:
			entries = append(entries, e.FileStrings(fullpath)...)
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

func (f *Folder) ReadFrom(path string) error {
	dirs, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range dirs {
		if entry.IsDir() {
			folder := f.Folder(entry.Name(), func(f *Folder) {})
			err = folder.ReadFrom(path + "/" + entry.Name())
			if err != nil {
				return err
			}
		} else if entry.Type().IsRegular() {
			content, err := os.ReadFile(path + "/" + entry.Name())
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
			target, err := os.Readlink(path + "/" + entry.Name())
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
	return false
}

func (f *Folder) EqualWithReason(entry FolderEntry) (bool, op.Reason) {
	// TODO: deal with MODE
	return false, op.Reason{}
}

func (f *Folder) HasContent() bool {
	return false
}

func (f *Folder) Content() []byte {
	return []byte{}
}

func (f *Folder) Diff(b *Folder) []op.Operation {
	return Diff(f, b, true)
}

func (f *Folder) CaseInsensitiveDiff(b *Folder) []op.Operation {
	return Diff(f, b, false)
}

func (f *Folder) ContentString() string {
	panic("fsdt:folder.go(Folder does not implement contentString)")
}
