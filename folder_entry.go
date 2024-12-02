package fsdt

// Define a custom type for FolderEntryType
type FolderEntryType string

// Define constants for FolderEntryType
const (
	FOLDER   FolderEntryType = "folder"
	FILE     FolderEntryType = "file"
	SYMLINK  FolderEntryType = "symlink"
	HARDLINK FolderEntryType = "hardlink"
)

type FolderEntry interface {
	WriteTo(location string) error
	Clone() FolderEntry
	Strings(prefix string) []string
	RemoveOperation(relativePath string) Operation
	CreateOperation(relativePath string) Operation
	Type() FolderEntryType
	Equal(FolderEntry) bool
	EqualWithReason(FolderEntry) (bool, Reason)
	HasContent() bool
	Content() []byte
	ContentString() string
}
