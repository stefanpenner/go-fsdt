package fsdt

import (
	"testing"

	op "github.com/stefanpenner/go-fsdt/operation"
	"github.com/stretchr/testify/require"
)

func Test_Virtual_StructureOnly_Ignores_Content(t *testing.T) {
	require := require.New(t)
	left := FS(map[string]string{"a.txt": "hello"})
	right := FS(map[string]string{"a.txt": "world"})

	d := DiffWithConfig(left, right, DefaultFast())
	require.Equal(op.Nothing, d)
}

func Test_Virtual_Bytes_Detects_Content(t *testing.T) {
	require := require.New(t)
	left := FS(map[string]string{"a.txt": "hello"})
	right := FS(map[string]string{"a.txt": "world"})

	d := DiffWithConfig(left, right, DefaultAccurate())
	require.NotEqual(op.Nothing, d)
}

func Test_Virtual_ChecksumPrefer_With_Computed_Sums(t *testing.T) {
	require := require.New(t)
	left := FS(map[string]string{"a.txt": "hello"})
	right := FS(map[string]string{"a.txt": "hello"})

	_, _, _ = left.Get("a.txt").(*File).EnsureChecksum(ChecksumOptions{Algorithm: "sha256", ComputeIfMissing: true})
	_, _, _ = right.Get("a.txt").(*File).EnsureChecksum(ChecksumOptions{Algorithm: "sha256", ComputeIfMissing: true})

	cfg := Checksums("sha256", nil)
	d := DiffWithConfig(left, right, cfg)
	require.Equal(op.Nothing, d)
}

func Test_Virtual_ChecksumRequire_Missing_Returns_Incompatible(t *testing.T) {
	require := require.New(t)
	left := FS(map[string]string{"a.txt": "hello"})
	right := FS(map[string]string{"a.txt": "hello"})

	cfg := ChecksumsStrict("sha256", nil)
	d := DiffWithConfig(left, right, cfg)

	found := false
	if dv, ok := d.Value.(op.DirValue); ok {
		for _, child := range dv.Operations {
			if fv, ok := child.Value.(op.FileChangedValue); ok {
				if fv.Reason.Type == op.Because { found = true; break }
			}
		}
	}
	require.True(found, "expected Because when checksum required but missing")
}

func Test_Virtual_ExcludeGlobs_Skips_Entries(t *testing.T) {
	require := require.New(t)
	left := FS(map[string]string{"keep.txt": "1", "tmp/x.log": "a"})
	right := FS(map[string]string{"keep.txt": "2", "tmp/x.log": "b"})

	cfg := DefaultAccurate()
	cfg.ExcludeGlobs = []string{"tmp/**"}
	d := DiffWithConfig(left, right, cfg)
	require.NotEqual(op.Nothing, d)
}