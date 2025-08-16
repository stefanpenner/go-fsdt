package fsdt

import (
	"archive/tar"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func Benchmark_Traversal(b *testing.B) {
	dir := b.TempDir()
	root := filepath.Join(dir, "root")
	createDeepFiles(root, 3, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		folder := NewFolder()
		_ = folder.ReadFromWithOptions(root, LoadOptions{})
	}
}

func Benchmark_Diff_Basic(b *testing.B) {
	dir := b.TempDir()
	rootA := filepath.Join(dir, "a")
	rootB := filepath.Join(dir, "b")
	createDeepFiles(rootA, 3, 5)
	copyTree(rootA, rootB)

	a := NewFolder()
	bld := NewFolder()
	_ = a.ReadFromWithOptions(rootA, LoadOptions{})
	_ = bld.ReadFromWithOptions(rootB, LoadOptions{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Diff(a, bld, true)
	}
}

func Benchmark_Hash_NoXAttr(b *testing.B) {
	dir := b.TempDir()
	root := filepath.Join(dir, "root")
	createDeepFiles(root, 3, 5)

	folder := NewFolder()
	_ = folder.ReadFromWithOptions(root, LoadOptions{})

	opts := ChecksumOptions{Algorithm: "sha256", ComputeIfMissing: true, StreamFromDiskIfAvailable: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		folder.checksum = nil
		_, _, _ = folder.EnsureChecksum(opts)
	}
}

func Benchmark_Hash_WithSidecar(b *testing.B) {
	dir := b.TempDir()
	root := filepath.Join(dir, "root")
	sidecar := filepath.Join(dir, "sidecar")
	createDeepFiles(root, 3, 5)

	folder := NewFolder()
	_ = folder.ReadFromWithOptions(root, LoadOptions{})

	// prepopulate sidecar
	pre := ChecksumOptions{Algorithm: "sha256", ComputeIfMissing: true, StreamFromDiskIfAvailable: true, SidecarDir: sidecar, RootPath: root}
	_, _, _ = folder.EnsureChecksum(pre)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// simulate fresh read
		folder2 := NewFolder()
		_ = folder2.ReadFromWithOptions(root, LoadOptions{})
		_, _, _ = folder2.EnsureChecksum(ChecksumOptions{Algorithm: "sha256", ComputeIfMissing: true, StreamFromDiskIfAvailable: true, SidecarDir: sidecar, RootPath: root})
	}
}

// Benchmark untarring with and without checksum derivation, using medium-size files.
func Benchmark_Untar_WithAndWithoutChecksum(b *testing.B) {
	// build an in-memory tar with several files
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	files := map[string]int{
		"a.bin": 256 * 1024,
		"b.bin": 512 * 1024,
		"c.bin": 1024 * 1024,
	}
	payload := bytes.Repeat([]byte{0xAB}, 1024)
	for name, size := range files {
		// file header
		hdr := &tar.Header{Name: name, Mode: 0644, Size: int64(size)}
		if err := tw.WriteHeader(hdr); err != nil { b.Fatal(err) }
		remaining := size
		for remaining > 0 {
			chunk := payload
			if remaining < len(chunk) { chunk = chunk[:remaining] }
			if _, err := tw.Write(chunk); err != nil { b.Fatal(err) }
			remaining -= len(chunk)
		}
	}
	_ = tw.Close()

	data := buf.Bytes()

	b.Run("no-checksum", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := ReadFromTarReader(bytes.NewReader(data))
			if err != nil { b.Fatal(err) }
		}
	})
	b.Run("checksum-sha256", func(b *testing.B) {
		opts := TarReadOptions{ComputeFileChecksum: true, ChecksumAlgorithm: "sha256"}
		for i := 0; i < b.N; i++ {
			_, err := ReadFromTarReaderWithOptions(bytes.NewReader(data), opts)
			if err != nil { b.Fatal(err) }
		}
	})
}

// helpers
func createDeepFiles(root string, depth, width int) {
	_ = os.MkdirAll(root, 0755)
	for d := 0; d < depth; d++ {
		base := filepath.Join(root, fmt.Sprintf("d_%d", d))
		_ = os.MkdirAll(base, 0755)
		for i := 0; i < width; i++ {
			_ = os.WriteFile(filepath.Join(base, fmt.Sprintf("f_%d.txt", i)), []byte(fmt.Sprintf("%d-%d", d, i)), 0644)
		}
	}
}

func copyTree(src, dst string) {
	_ = filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(src, path)
		tgt := filepath.Join(dst, rel)
		if d.IsDir() {
			_ = os.MkdirAll(tgt, 0755)
			return nil
		}
		data, _ := os.ReadFile(path)
		_ = os.MkdirAll(filepath.Dir(tgt), 0755)
		_ = os.WriteFile(tgt, data, 0644)
		return nil
	})
}