package operation

import "fmt"

// FileValue represents the value for file operations
type FileValue struct {
	Content     []byte
	Mode        uint32
	Permissions uint32
	Size        int64
	Checksum    string
}

// String returns a string representation of the file value
func (f FileValue) String() string {
	if f.Size > 0 {
		return fmt.Sprintf("file (%d bytes, mode: %o)", f.Size, f.Mode)
	}
	return fmt.Sprintf("file (mode: %o)", f.Mode)
}

// FileChangedValue represents the value for file change operations
type FileChangedValue struct {
	Before FileValue
	After  FileValue
	Reason *Reason
}

// String returns a string representation of the file change value
func (f FileChangedValue) String() string {
	return fmt.Sprintf("changed from %s to %s", f.Before.String(), f.After.String())
}

// NewCreateFileOperation creates a new file creation operation
func NewCreateFileOperation(relativePath string, content []byte, mode uint32) Operation {
	value := FileValue{
		Content: content,
		Mode:    mode,
		Size:    int64(len(content)),
	}

	reason := &Reason{
		Type:    ReasonNew,
		Message: "File does not exist in target",
	}

	return NewOperation(relativePath, CreateFile, value, reason)
}

// NewRemoveFileOperation creates a new file removal operation
func NewRemoveFileOperation(relativePath string) Operation {
	reason := &Reason{
		Type:    ReasonMissing,
		Message: "File exists in source but not in target",
	}

	return NewOperation(relativePath, RemoveFile, nil, reason)
}

// NewChangeFileOperation creates a new file change operation
func NewChangeFileOperation(relativePath string, before, after FileValue, changeType ReasonType) Operation {
	value := FileChangedValue{
		Before: before,
		After:  after,
	}

	reason := &Reason{
		Type:    changeType,
		Before:  before,
		After:   after,
		Message: fmt.Sprintf("File %s changed", changeType),
	}

	return NewOperation(relativePath, ChangeFile, value, reason)
}
