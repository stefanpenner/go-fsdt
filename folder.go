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
	// optional exclude globs
	excludeGlobs []string
	// folder-level checksum
	checksum          []byte
	checksumAlgorithm string
	// source path on disk for streaming/xattr
	sourcePath string
	// folder-level checksum policy
	policy ChecksumPolicy
}

var DEFAULT_FOLDER_MODE = os.FileMode(os.ModeDir | 0755)

func NewFolder(cb ...func(f *Folder)) *Folder {
	folder := &Folder{
		_entries: map[string]FolderEntry{},
		mode:     DEFAULT_FOLDER_MODE,
		policy:   ChecksumPolicy{Strategy: ChecksumOff},
	}

	for _, cb := range cb {
		cb(folder)
	}

	return folder
}

// Put inserts or replaces an entry under the given name.
func (f *Folder) Put(name string, entry FolderEntry) {
	f._entries[name] = entry
}

func (f *Folder) SetExcludeGlobs(globs []string) {
	f.excludeGlobs = append([]string(nil), globs...)
}

func (f *Folder) ExcludeGlobs() []string {
	return append([]string(nil), f.excludeGlobs...)
}

func (f *Folder) SetChecksum(algorithm string, digest []byte) {
	f.checksumAlgorithm = algorithm
	f.checksum = append([]byte(nil), digest...)
}

// EnsureChecksum ensures this folder has a checksum; computes from children if missing and optionally writes xattr
func (f *Folder) EnsureChecksum(opts ChecksumOptions) ([]byte, string, bool) {
	if d, n, ok := f.Checksum(); ok {
		if f.sourcePath != "" {
			writeChecksumCache(f.sourcePath, d, opts)
		}
		return d, n, true
	}
	if !opts.ComputeIfMissing || opts.Algorithm == "" {
		return nil, "", false
	}
	d := computeFolderChecksum(f, opts.Algorithm)
	if d == nil {
		return nil, "", false
	}
	f.SetChecksum(opts.Algorithm, d)
	if f.sourcePath != "" {
		writeChecksumCache(f.sourcePath, d, opts)
	}
	return d, opts.Algorithm, true
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
	clone.excludeGlobs = append([]string(nil), f.excludeGlobs...)
	clone.checksum = append([]byte(nil), f.checksum...)
	clone.checksumAlgorithm = f.checksumAlgorithm
	clone.sourcePath = f.sourcePath
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
	error := folder.ReadFromWithOptions(path, LoadOptions{})
	return folder, error
}

// LoadOptions controls how filesystem metadata such as xattr checksums are loaded.
type LoadOptions struct {
	// If set, attempt to read checksum from xattr with this key (e.g., "user.sha256")
	XAttrChecksumKey string
	// Label to store with the checksum so algorithms can be matched during compare, e.g., "sha256"
	ChecksumAlgorithm string
	// If true and no xattr is present, compute checksum from content using the provided algorithm
	ComputeChecksumIfMissing bool
	// If true, write any computed checksums back to xattr
	WriteComputedChecksumToXAttr bool
	// If true, do not load file bytes into memory; size and sourcePath are recorded for streaming/lazy ops
	SkipContentRead bool
	// If true, compute a folder-level checksum (from self + children) when missing
	ComputeFolderChecksumIfMissing bool
	// If true, write computed folder checksum back to xattr when missing
	WriteComputedFolderChecksumToXAttr bool
}

func (f *Folder) ReadFrom(path string) error {
	return f.ReadFromWithOptions(path, LoadOptions{})
}

func (f *Folder) ReadFromWithOptions(path string, opts LoadOptions) error {
	f.sourcePath = path
	dirs, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range dirs {
		if entry.IsDir() {
			folder := f.Folder(entry.Name(), func(f *Folder) {})
			err = folder.ReadFromWithOptions(filepath.Join(path, entry.Name()), opts)
			if err != nil {
				return err
			}
		} else if entry.Type().IsRegular() {
			full := filepath.Join(path, entry.Name())
			info, err := entry.Info()
			if err != nil { return err }
			var content []byte
			if !opts.SkipContentRead {
				data, err := os.ReadFile(full)
				if err != nil { return err }
				content = data
			}
			file := f.File(entry.Name(), FileOptions{ Content: content, Mode: info.Mode(), MTime: info.ModTime(), Size: info.Size() })
			file.sourcePath = full

			if opts.XAttrChecksumKey != "" {
				if digest, ok, _ := readXAttrChecksum(full, opts.XAttrChecksumKey); ok {
					file.SetChecksum(opts.ChecksumAlgorithm, digest)
				} else if opts.ComputeChecksumIfMissing && opts.ChecksumAlgorithm != "" {
					d := computeChecksumFromPathOrBytes(opts.ChecksumAlgorithm, full, content)
					if d != nil {
						file.SetChecksum(opts.ChecksumAlgorithm, d)
						if opts.WriteComputedChecksumToXAttr {
							_ = writeXAttrChecksum(full, opts.XAttrChecksumKey, d)
						}
					}
				}
			}
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

	// Compute folder checksum if requested
	if opts.ChecksumAlgorithm != "" && opts.ComputeFolderChecksumIfMissing {
		if _, _, has := f.Checksum(); !has {
			d := computeFolderChecksum(f, opts.ChecksumAlgorithm)
			if d != nil {
				f.SetChecksum(opts.ChecksumAlgorithm, d)
				if opts.WriteComputedFolderChecksumToXAttr && opts.XAttrChecksumKey != "" && f.sourcePath != "" {
					_ = writeXAttrChecksum(f.sourcePath, opts.XAttrChecksumKey, d)
				}
			}
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

// Checksum for folders returns folder-level checksum
func (f *Folder) Checksum() ([]byte, string, bool) {
	if len(f.checksum) == 0 {
		return nil, "", false
	}
	return append([]byte(nil), f.checksum...), f.checksumAlgorithm, true
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

// computeFolderChecksum computes a folder checksum as a digest of folder metadata and child checksums.
func computeFolderChecksum(folder *Folder, algorithm string) []byte {
	h := newHash(algorithm)
	if h == nil {
		return nil
	}
	// include folder mode
	ioWriteString(h, fmt.Sprintf("dir|mode:%o\n", folder.mode))
	// include each child (sorted by name)
	for _, name := range folder.Entries() {
		if shouldExclude(normalizePath("", name), folder.excludeGlobs) {
			continue
		}
		entry := folder._entries[name]
		// for files and folders, prefer their checksum; compute if file has none
		switch e := entry.(type) {
		case *File:
			if d, n, ok := e.Checksum(); ok {
				ioWriteString(h, fmt.Sprintf("file|%s|algo:%s|%x\n", name, n, d))
			} else {
				// compute from content or path
				d := computeChecksumFromPathOrBytes(algorithm, e.sourcePath, e.content)
				ioWriteString(h, fmt.Sprintf("file|%s|algo:%s|%x\n", name, algorithm, d))
			}
		case *Folder:
			if d, n, ok := e.Checksum(); ok {
				ioWriteString(h, fmt.Sprintf("dir|%s|algo:%s|%x\n", name, n, d))
			} else {
				d := computeFolderChecksum(e, algorithm)
				ioWriteString(h, fmt.Sprintf("dir|%s|algo:%s|%x\n", name, algorithm, d))
			}
		case *Link:
			ioWriteString(h, fmt.Sprintf("link|%s|%s\n", name, e.Target()))
		default:
			ioWriteString(h, fmt.Sprintf("unknown|%s\n", name))
		}
	}
	return h.Sum(nil)
}
