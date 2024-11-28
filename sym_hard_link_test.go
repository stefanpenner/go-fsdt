package fsdt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSymlink(t *testing.T) {
	assert := assert.New(t)

	folder := NewFolder()

	folder.Symlink("bar", "nowhere")                // broken
	folder.Symlink("baz", "somewhere")              // somewhere
	folder.Symlink("baz", "../somewhere-else")      // outside
	folder.Symlink("baz", "../<current location>/") // cycle

	// test right entry
	// diffing
	// write to disk
	// test disk
	// restore from disk
	// test right entry
}

func TestHardlink(t *testing.T) {
	assert := assert.New(t)

	folder := NewFolder()

	folder.Hardlink("bar", "nowhere")                // broken
	folder.Hardlink("baz", "somewhere")              // somewhere
	folder.Hardlink("baz", "../somewhere-else")      // outside
	folder.Hardlink("baz", "../<current location>/") // cycle

	// test right entry
	// diffing
	// write to disk
	// test disk
	// restore from disk
	// test right entry
}
