package fsdt

import (
	"testing"
	"time"

	op "github.com/stefanpenner/go-fsdt/operation"
	"github.com/stretchr/testify/require"
)

func Test_MTimeExcluded_With_Config(t *testing.T) {
	require := require.New(t)

	left := NewFolder()
	right := NewFolder()

	// Same content and size, different mtime
	t1 := time.Unix(1000, 0)
	t2 := time.Unix(2000, 0)

	left.File("a.txt", FileOptions{Content: []byte("same"), Mode: DEFAULT_FILE_MODE, MTime: t1})
	right.File("a.txt", FileOptions{Content: []byte("same"), Mode: DEFAULT_FILE_MODE, MTime: t2})

	// Accurate default compares mtime => change
	dAcc := DiffWithConfig(left, right, DefaultAccurate())
	require.NotEqual(op.Nothing, dAcc)

	// Accurate but with mtime disabled => no change
	cfg := DefaultAccurateNoMTime()
	dNoMTime := DiffWithConfig(left, right, cfg)
	require.Equal(op.Nothing, dNoMTime)
}