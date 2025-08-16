package fsdt

import (
	"bytes"
	"os"
	"time"

	op "github.com/stefanpenner/go-fsdt/operation"
)

type File struct {
	// TODO: FileInfo might be handy
	content []byte
	mode    os.FileMode
	checksum        []byte
	checksumAlgorithm string
	mtime   time.Time
	size    int64
	sourcePath string
}

type FileOptions struct {
	Content []byte
	Mode    os.FileMode
	// Optional checksum metadata, e.g. xattr-provided digest
	Checksum          []byte
	ChecksumAlgorithm string
	// Optional file metadata
	MTime time.Time
	Size  int64
}

var DEFAULT_FILE_MODE = os.FileMode(0644)

func NewFileString(content string) *File {
	return NewFile(FileOptions{Content: []byte(content)})
}

func NewFile(content ...FileOptions) *File {
	if len(content) == 0 {
		return &File{
			content: []byte{},
			mode:    DEFAULT_FILE_MODE,
		}
	}

	opts := content[0]
	mode := opts.Mode
	if mode == 0 {
		mode = DEFAULT_FILE_MODE
	}
	computedSize := opts.Size
	if computedSize == 0 && opts.Content != nil {
		computedSize = int64(len(opts.Content))
	}
	return &File{
		content:            opts.Content,
		mode:               mode,
		checksum:          opts.Checksum,
		checksumAlgorithm: opts.ChecksumAlgorithm,
		mtime:             opts.MTime,
		size:              computedSize,
	}
}

func (f *File) Strings(prefix string) []string {
	return []string{prefix}
}

func (f *File) Clone() FolderEntry {
	return &File{
		content:            append([]byte(nil), f.content...),
		mode:               f.mode,
		checksum:          append([]byte(nil), f.checksum...),
		checksumAlgorithm: f.checksumAlgorithm,
		mtime:             f.mtime,
		size:              f.size,
		sourcePath:        f.sourcePath,
	}
}

func (f *File) RemoveOperation(relativePath string, reason op.Reason) op.Operation {
	// TODO: reason
	return op.NewUnlink(relativePath)
}

func (f *File) CreateOperation(relativePath string, reason op.Reason) op.Operation {
	// TODO: reason
	return op.NewFileOperation(relativePath)
}

func (f *File) ChangeOperation(relativePath string, reason op.Reason, operations ...op.Operation) op.Operation {
	return op.Operation{
		Operand:      op.ChangeFile,
		RelativePath: relativePath,
		Value: op.FileChangedValue{
			Reason: reason,
		},
	}
}

func (f *File) Type() FolderEntryType {
	return FILE
}

func (f *File) WriteTo(location string) error {
	file, err := os.OpenFile(location, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.mode)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(f.content)
	return err
}

func (f *File) Content() []byte {
	return f.content
}

func (f *File) Mode() os.FileMode {
	return f.mode
}

func (f *File) MTime() time.Time {
	return f.mtime
}

func (f *File) Size() int64 {
	return f.size
}

func (f *File) SourcePath() (string, bool) {
	if f.sourcePath == "" {
		return "", false
	}
	return f.sourcePath, true
}

func (f *File) ContentString() string {
	return string(f.content)
}

// EnsureChecksum makes sure a checksum is present for this file and optionally persists it to xattr.
func (f *File) EnsureChecksum(opts ChecksumOptions) ([]byte, string, bool) {
	if d, n, ok := f.Checksum(); ok {
		if path, has := f.SourcePath(); has {
			writeChecksumCache(path, d, opts)
		}
		return d, n, true
	}
	if !opts.ComputeIfMissing || opts.Algorithm == "" {
		return nil, "", false
	}
	var content []byte
	if opts.StreamFromDiskIfAvailable {
		if path, has := f.SourcePath(); has {
			if data, err := readAllStreaming(path); err == nil {
				content = data
			}
		}
	}
	if content == nil {
		content = f.content
	}
	d := computeChecksum(opts.Algorithm, content)
	if d == nil {
		return nil, "", false
	}
	f.SetChecksum(opts.Algorithm, d)
	if path, has := f.SourcePath(); has {
		writeChecksumCache(path, d, opts)
	}
	return d, opts.Algorithm, true
}

// Checksum returns the stored checksum digest (if any) and its algorithm name.
func (f *File) Checksum() ([]byte, string, bool) {
	if len(f.checksum) == 0 {
		return nil, "", false
	}
	return append([]byte(nil), f.checksum...), f.checksumAlgorithm, true
}

// SetChecksum sets checksum metadata for the file (e.g., from xattr) with the given algorithm name.
func (f *File) SetChecksum(algorithm string, digest []byte) {
	f.checksumAlgorithm = algorithm
	f.checksum = append([]byte(nil), digest...)
}

func (f *File) Equal(entry FolderEntry) bool {
	equal, _ := f.EqualWithReason(entry)
	return equal
}

func (f *File) EqualWithReason(entry FolderEntry) (bool, op.Reason) {
	file, isFile := entry.(*File)

	if isFile {
		if f.mode != file.mode {
			return false, op.Reason{
				Type:   op.ModeChanged,
				Before: f.Mode(),
				After:  file.Mode(),
			}
		}

		if bytes.Equal(f.content, file.content) {
			return true, op.Reason{}
		} else {
			// TODO: maybe should show offset and first char difference
			return false, op.Reason{
				Type:   op.ContentChanged,
				Before: f.content,
				After:  file.content,
			}
		}
	}
	return false, op.Reason{
		Type:   op.TypeChanged,
		Before: f.Type(),
		After:  entry.Type(),
	}
}

func (f *File) HasContent() bool {
	return true
}
