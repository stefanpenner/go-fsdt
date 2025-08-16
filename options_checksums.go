package fsdt

type ChecksumOptions struct {
	Algorithm string
	XAttrKey  string
	ComputeIfMissing bool
	WriteToXAttr     bool
	StreamFromDiskIfAvailable bool
}

// EnsureEntryChecksum ensures the entry has a checksum according to options and returns it.
// For files/folders this may compute and optionally persist to xattr.
func EnsureEntryChecksum(e FolderEntry, opts ChecksumOptions) ([]byte, string, bool) {
	switch v := e.(type) {
	case *File:
		return v.EnsureChecksum(opts)
	case *Folder:
		return v.EnsureChecksum(opts)
	default:
		return nil, "", false
	}
}