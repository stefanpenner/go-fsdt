package fsdt

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSymlink(t *testing.T) {
	folder := NewFolder()

	folder.File("target_file", FileOptions{Content: []byte("hello")})
	folder.Folder("target_folder")
	folder.Symlink("link_to_file", "target_file")
	folder.Symlink("link_to_folder", "target_folder")

	// verify
	assert := assert.New(t)
	assert.Equal([]string{"link_to_file -> target_file", "link_to_folder -> target_folder", "target_file", "target_folder/"}, folder.Strings(""))
}

func TestSymlinkWriteAndRead(t *testing.T) {
	folder := FS(map[string]string{"target_file": "hello"})
	folder.Mk("target_folder")
	folder.Symlink("link_to_file", "target_file")
	folder.Symlink("link_to_folder", "target_folder")

	location := filepath.Join(t.TempDir(), "folder")
	require.NoError(t, folder.WriteTo(location))

	loadedFolder := NewFolder()
	require.NoError(t, loadedFolder.ReadFrom(location))

	assert := assert.New(t)
	assert.Equal("hello", loadedFolder.Get("target_file").ContentString())
	assert.Equal("target_file", loadedFolder.Get("link_to_file").(*Link).Target())
	assert.Equal("target_folder", loadedFolder.Get("link_to_folder").(*Link).Target())
}
