package fsdt

import (
	"encoding/hex"
	"os"
	"path/filepath"
)

type ChecksumOptions struct {
	Algorithm string
	XAttrKey  string
	ComputeIfMissing bool
	WriteToXAttr     bool
	StreamFromDiskIfAvailable bool
	// Optional sidecar cache config
	SidecarDir string
	RootPath   string
}

// EnsureEntryChecksum ensures the entry has a checksum according to options and returns it.
// For files/folders this may compute and optionally persist to xattr/sidecar.
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

func writeChecksumCache(path string, digest []byte, opts ChecksumOptions) {
	if len(digest) == 0 {
		return
	}
	if opts.WriteToXAttr && opts.XAttrKey != "" {
		_ = writeXAttrChecksum(path, opts.XAttrKey, digest)
	}
	if opts.SidecarDir != "" && opts.RootPath != "" {
		rel, err := filepath.Rel(opts.RootPath, path)
		if err == nil {
			outPath := filepath.Join(opts.SidecarDir, rel+"."+opts.Algorithm)
			_ = os.MkdirAll(filepath.Dir(outPath), 0755)
			_ = os.WriteFile(outPath, []byte(hex.EncodeToString(digest)), 0644)
		}
	}
}