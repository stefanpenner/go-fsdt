package fsdt

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ReadFromTarReader constructs a Folder from a tar stream (optionally gzip-compressed if caller wraps r).
func ReadFromTarReader(r io.Reader) (*Folder, error) {
	tr := tar.NewReader(r)
	root := NewFolder()

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		name := path.Clean(hdr.Name)
		if name == "." || name == "" {
			continue
		}

		// Split path using forward slashes (tar format standard)
		components := strings.Split(name, "/")
		parent := ensureFolder(root, components[:len(components)-1])
		base := components[len(components)-1]

		mode := os.FileMode(hdr.Mode)
		switch hdr.Typeflag {
		case tar.TypeDir:
			// Explicit directory entry
			folder := parent.Folder(base)
			if mode != 0 { folder.mode = (mode | os.ModeDir) }
		case tar.TypeSymlink:
			parent.Symlink(base, hdr.Linkname)
		case tar.TypeReg, tar.TypeRegA:
			// Read file content
			var buf []byte
			if hdr.Size > 0 {
				buf = make([]byte, hdr.Size)
				if _, err := io.ReadFull(tr, buf); err != nil {
					return nil, err
				}
			}
			file := parent.File(base, FileOptions{Content: buf, Mode: mode, MTime: hdr.ModTime, Size: hdr.Size})
			_ = file // silence linters if not using fields immediately
		case tar.TypeLink:
			// Hard links are not supported in fsdt
			return nil, errors.New("tar: hard links not supported")
		default:
			// Skip other types (fifo, char, block, etc.)
			continue
		}
	}
	return root, nil
}

// ReadFromTarFile opens a .tar or .tar.gz file and returns a Folder.
func ReadFromTarFile(path string) (*Folder, error) {
	f, err := os.Open(path)
	if err != nil { return nil, err }
	defer f.Close()

	if isGzipTar(path) {
		gzr, err := gzip.NewReader(f)
		if err != nil { return nil, err }
		defer gzr.Close()
		return ReadFromTarReader(gzr)
	}
	return ReadFromTarReader(f)
}

// WriteTarTo writes the Folder contents into a tar stream.
func (f *Folder) WriteTarTo(w io.Writer) error {
			tw := tar.NewWriter(w)
			defer tw.Close()
			return writeFolderToTar(tw, f, "")
}

// WriteToTarFile writes a .tar or .tar.gz depending on the file extension.
func (f *Folder) WriteToTarFile(p string) error {
	file, err := os.Create(p)
	if err != nil { return err }
	defer file.Close()

	if isGzipTar(p) {
		gz := gzip.NewWriter(file)
		defer gz.Close()
		return f.WriteTarTo(gz)
	}
	return f.WriteTarTo(file)
}

func writeFolderToTar(tw *tar.Writer, folder *Folder, prefix string) error {
	// Ensure directory header for non-root folders to preserve modes
	if prefix != "" {
		hdr := &tar.Header{
			Name:     toTarPath(prefix) + "/",
			Typeflag: tar.TypeDir,
			Mode:     int64(folder.mode.Perm()),
		}
		if err := tw.WriteHeader(hdr); err != nil { return err }
	}
	for _, name := range folder.Entries() {
		entry := folder.Get(name)
		full := name
		if prefix != "" { full = prefix + "/" + name }
		switch e := entry.(type) {
		case *Folder:
			if err := writeFolderToTar(tw, e, full); err != nil { return err }
		case *File:
			hdr := &tar.Header{
				Name:     toTarPath(full),
				Typeflag: tar.TypeReg,
				Mode:     int64(e.mode.Perm()),
				Size:     int64(len(e.content)),
				ModTime:  e.mtime,
			}
			if err := tw.WriteHeader(hdr); err != nil { return err }
			if len(e.content) > 0 {
				if _, err := tw.Write(e.content); err != nil { return err }
			}
		case *Link:
			hdr := &tar.Header{
				Name:     toTarPath(full),
				Typeflag: tar.TypeSymlink,
				Linkname: e.Target(),
				Mode:     0777,
			}
			if err := tw.WriteHeader(hdr); err != nil { return err }
		default:
			return errors.New("unexpected entry type while writing tar")
		}
	}
	return nil
}

func ensureFolder(root *Folder, components []string) *Folder {
	cur := root
	for _, part := range components {
		if part == "" { continue }
		// Ensure child exists and is a folder
		child, ok := cur._entries[part]
		if !ok {
			child = NewFolder()
			cur._entries[part] = child
		}
		fchild, isFolder := child.(*Folder)
		if !isFolder {
			// Replace non-folder with folder (last write wins in tar ordering)
			fchild = NewFolder()
			cur._entries[part] = fchild
		}
		cur = fchild
	}
	return cur
}

func isGzipTar(p string) bool {
	lower := strings.ToLower(p)
	return strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz")
}

func toTarPath(p string) string { return strings.ReplaceAll(filepath.ToSlash(p), "\\", "/") }