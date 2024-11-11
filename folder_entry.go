package fs

// Define a custom type for FolderEntryType
type FolderEntryType string

// Define constants for FolderEntryType
const (
	FOLDER FolderEntryType = "folder"
	FILE   FolderEntryType = "file"
)

type FolderEntry interface {
	WriteTo(location string) error
	Clone() FolderEntry
	Strings(prefix string) []string
	RemoveOperation(relativePath string) Operation
	CreateOperation(relativePath string) Operation
	Type() FolderEntryType
	Equal(FolderEntry) bool
	HasContent() bool
	Content() []byte
	ContentString() string
}
