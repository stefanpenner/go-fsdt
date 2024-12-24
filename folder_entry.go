package fsdt

import op "github.com/stefanpenner/go-fsdt/operation"

// Define a custom type for FolderEntryType
type FolderEntryType string

// Define constants for FolderEntryType
const (
	FOLDER   FolderEntryType = "folder"
	FILE     FolderEntryType = "file"
	SYMLINK  FolderEntryType = "symlink"
	HARDLINK FolderEntryType = "hardlink" // Not really supported currently, may never choose to support them.
)

type FolderEntry interface {
	WriteTo(location string) error
	Clone() FolderEntry
	Strings(prefix string) []string
	// TODO: consider all operations taking optional reason
	RemoveOperation(relativePath string) op.Operation
	CreateOperation(relativePath string) op.Operation
	Type() FolderEntryType
	Equal(FolderEntry) bool
	EqualWithReason(FolderEntry) (bool, op.Reason)
	HasContent() bool
	Content() []byte
	ContentString() string
}
