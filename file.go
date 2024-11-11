package fs

import (
	"bytes"
	"os"
)

type File struct {
	// FileInfo might be handy
	content []byte
	// encoding string
}

type FileOptions struct {
	Content []byte
	Perm    os.FileMode
}

func NewFileString(content string) *File {
	return NewFile(FileOptions{Content: []byte(content)})
}

func NewFile(content ...FileOptions) *File {
	if len(content) == 0 {
		return &File{
			content: []byte{},
			// encoding: "utf-8",
		}
	}

	return &File{
		content: content[0].Content,
		// encoding: "utf-8",
	}
}

func (f *File) Strings(prefix string) []string {
	return []string{prefix}
}

func (f *File) Clone() FolderEntry {
	return &File{
		content: append([]byte(nil), f.content...),
		// encoding: f.encoding,
	}
}

func (f *File) RemoveOperation(relativePath string) Operation {
	return NewUnlink(relativePath)
}

func (f *File) CreateOperation(relativePath string) Operation {
	return NewCreate(relativePath)
}

func (f *File) Type() FolderEntryType {
	return FILE
}

func (f *File) WriteTo(location string) error {
	file, err := os.Create(location)
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

func (f *File) ContentString() string {
	return string(f.content)
}

func (f *File) Equal(entry FolderEntry) bool {
	if file, ok := entry.(*File); ok {
		return bytes.Equal(f.content, file.content)
	}
	return false
}

func (f *File) HasContent() bool {
	return true
}
