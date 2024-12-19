package fsdt

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileMode(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	file := NewFile()

	assert.Equal(file.Mode(), os.FileMode(0644))
	assert.Equal(file.Mode().Perm(), os.FileMode(0644))
	tempdir := t.TempDir()

	file.WriteTo(tempdir + "/foo.txt")

	stat, err := os.Stat(tempdir + "/foo.txt")
	require.NoError(err)

	assert.Equal(os.FileMode(0644), stat.Mode())
	assert.Equal(os.FileMode(0644), stat.Mode().Perm())
}
