package fsdt

import (
	"os"
	"path/filepath"
	"testing"

	op "github.com/stefanpenner/go-fsdt/operation"
	"github.com/stretchr/testify/require"
)

func Test_FileChecksum_XAttr_And_Sidecar(t *testing.T) {
	require := require.New(t)
	dir := t.TempDir()
	root := filepath.Join(dir, "root")
	require.NoError(os.MkdirAll(root, 0755))
	dataPath := filepath.Join(root, "a.txt")
	require.NoError(os.WriteFile(dataPath, []byte("hello"), 0644))

	folder := NewFolder()
	require.NoError(folder.ReadFromWithOptions(root, LoadOptions{}))
	file := folder.Get("a.txt").(*File)

	opts := ChecksumOptions{
		Algorithm: "sha256",
		XAttrKey:  "user.sha256",
		ComputeIfMissing: true,
		WriteToXAttr: true,
		StreamFromDiskIfAvailable: true,
		SidecarDir: filepath.Join(dir, "sidecar"),
		RootPath:   root,
	}

	d, algo, ok := file.EnsureChecksum(opts)
	require.True(ok)
	require.Equal("sha256", algo)
	require.NotEmpty(d)

	// Check sidecar exists
	_, err := os.Stat(filepath.Join(opts.SidecarDir, "a.txt.sha256"))
	require.NoError(err)
}

func Test_FolderChecksum_Composed_From_Children(t *testing.T) {
	require := require.New(t)
	dir := t.TempDir()
	root := filepath.Join(dir, "root")
	require.NoError(os.MkdirAll(filepath.Join(root, "sub"), 0755))
	require.NoError(os.WriteFile(filepath.Join(root, "a.txt"), []byte("hello"), 0644))
	require.NoError(os.WriteFile(filepath.Join(root, "sub", "b.txt"), []byte("world"), 0644))

	folder := NewFolder()
	require.NoError(folder.ReadFromWithOptions(root, LoadOptions{}))

	// Ensure children have checksums
	_, _, _ = folder.Get("a.txt").(*File).EnsureChecksum(ChecksumOptions{Algorithm: "sha256", ComputeIfMissing: true, StreamFromDiskIfAvailable: true})
	sub := folder.Get("sub").(*Folder)
	_, _, _ = sub.Get("b.txt").(*File).EnsureChecksum(ChecksumOptions{Algorithm: "sha256", ComputeIfMissing: true, StreamFromDiskIfAvailable: true})

	d, algo, ok := folder.EnsureChecksum(ChecksumOptions{Algorithm: "sha256", ComputeIfMissing: true})
	require.True(ok)
	require.Equal("sha256", algo)
	require.NotEmpty(d)
}

func Test_FileChecksum_StreamVsMemory(t *testing.T) {
	require := require.New(t)
	dir := t.TempDir()
	root := filepath.Join(dir, "root")
	require.NoError(os.MkdirAll(root, 0755))
	dataPath := filepath.Join(root, "big.bin")
	// create ~1MB file
	buf := make([]byte, 1024*1024)
	for i := range buf {
		buf[i] = byte(i % 251)
	}
	require.NoError(os.WriteFile(dataPath, buf, 0644))

	folder := NewFolder()
	require.NoError(folder.ReadFromWithOptions(root, LoadOptions{}))
	file := folder.Get("big.bin").(*File)

	// Streaming
	d1, _, ok1 := file.EnsureChecksum(ChecksumOptions{Algorithm: "sha256", ComputeIfMissing: true, StreamFromDiskIfAvailable: true})
	require.True(ok1)
	// Memory
	file.checksum = nil
	d2, _, ok2 := file.EnsureChecksum(ChecksumOptions{Algorithm: "sha256", ComputeIfMissing: true, StreamFromDiskIfAvailable: false})
	require.True(ok2)
	require.Equal(d1, d2)
}

func Test_ExcludeGlobs_Diff_And_FolderChecksum(t *testing.T) {
	require := require.New(t)
	dir := t.TempDir()
	rootA := filepath.Join(dir, "a")
	rootB := filepath.Join(dir, "b")
	_ = os.MkdirAll(filepath.Join(rootA, "tmp"), 0755)
	_ = os.MkdirAll(filepath.Join(rootB, "tmp"), 0755)
	_ = os.WriteFile(filepath.Join(rootA, "keep.txt"), []byte("1"), 0644)
	_ = os.WriteFile(filepath.Join(rootB, "keep.txt"), []byte("2"), 0644)
	_ = os.WriteFile(filepath.Join(rootA, "tmp", "x.log"), []byte("noise"), 0644)
	_ = os.WriteFile(filepath.Join(rootB, "tmp", "x.log"), []byte("noise"), 0644)

	a := NewFolder()
	b := NewFolder()
	require.NoError(a.ReadFromWithOptions(rootA, LoadOptions{}))
	require.NoError(b.ReadFromWithOptions(rootB, LoadOptions{}))

	cfg := DefaultAccurate()
	cfg.ExcludeGlobs = []string{"tmp/**"}
	d := DiffWithConfig(a, b, cfg)
	require.NotEqual(op.Nothing, d) // keep.txt differs

	// Different globs => incompatible
	cfgB := cfg
	cfgB.ExcludeGlobs = []string{"tmp/**", "other/**"}
	d2 := DiffWithConfig(a, b, cfgB)
	require.Equal(op.ChangeFolder, d2.Operand)

	// Folder checksum excludes tmp/**
	a.SetExcludeGlobs(cfg.ExcludeGlobs)
	b.SetExcludeGlobs(cfg.ExcludeGlobs)
	_, _, oka := a.EnsureChecksum(ChecksumOptions{Algorithm: "sha256", ComputeIfMissing: true})
	_, _, okb := b.EnsureChecksum(ChecksumOptions{Algorithm: "sha256", ComputeIfMissing: true})
	require.True(oka)
	require.True(okb)
}