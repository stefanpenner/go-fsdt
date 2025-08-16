package fsdt

import (
	"path/filepath"
	"testing"

	op "github.com/stefanpenner/go-fsdt/operation"
	"github.com/stretchr/testify/require"
)

func Test_Scenario_FastVsAccurate(t *testing.T) {
	require := require.New(t)
	left := NewFolder()
	right := NewFolder()

	left.FileString("a.txt", "hello")
	right.FileString("a.txt", "world")

	// Fast: structure-only + mode (content ignored) => no changes
	d1 := DiffWithConfig(left, right, DefaultFast())
	require.Equal(op.Nothing, d1)

	// Accurate: byte comparison => change detected
	d2 := DiffWithConfig(left, right, DefaultAccurate())
	require.NotEqual(op.Nothing, d2)
}

func Test_Scenario_ChecksumEnsure_With_Sidecar(t *testing.T) {
	require := require.New(t)
	dir := t.TempDir()
	root := filepath.Join(dir, "root")
	side := filepath.Join(dir, "cache")

	left := NewFolder()
	right := NewFolder()

	// Same structure, different content
	left.FileString("a.txt", "hello")
	right.FileString("a.txt", "world")

	store := MultiStore{Stores: []ChecksumStore{SidecarStore{BaseDir: side, Root: root, Algorithm: "sha256"}}}
	cfg := Checksums("sha256", store)
	cfg.Strategy = ChecksumEnsure

	d := DiffWithConfig(left, right, cfg)
	require.NotEqual(op.Nothing, d)
}

func Test_Scenario_ChecksumPrefer_FallbackToBytes_Equal(t *testing.T) {
	require := require.New(t)
	left := NewFolder()
	right := NewFolder()

	left.FileString("a.txt", "same content")
	right.FileString("a.txt", "same content")

	cfg := DefaultAccurate()
	// Switch to checksum-prefer but without any store/algorithm so it must fallback to bytes
	cfg.Strategy = ChecksumPrefer
	cfg.Algorithm = "" // ensures no checksum path

	d := DiffWithConfig(left, right, cfg)
	require.Equal(op.Nothing, d)
}

func Test_Scenario_ChecksumPrefer_FallbackToBytes_Different_ExplainIncludesLengths(t *testing.T) {
	require := require.New(t)
	left := NewFolder()
	right := NewFolder()

	left.FileString("a.txt", "aaaaaaaaaa")
	right.FileString("a.txt", "bbbbbbbbbb") // same length, different bytes

	cfg := DefaultAccurate()
	cfg.Strategy = ChecksumPrefer
	cfg.Algorithm = ""

	d := DiffWithConfig(left, right, cfg)
	require.NotEqual(op.Nothing, d)

	explain := op.Explain(d)
	require.Contains(explain, "content differs (len before 10, after 10)")
}