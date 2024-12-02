package fsdt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSymlink(t *testing.T) {
	folder := NewFolder()

	folder.File("target_file", FileOptions{Content: []byte("hello")})
	folder.Folder("target_folder")

	t.Run("basic symlink operations", func(t *testing.T) {
		assert := assert.New(t)
		folder := folder.Clone().(*Folder)

		link_to_file := folder.Symlink("link_to_file", "target_file")
		link_to_folder := folder.Symlink("link_to_folder", "target_folder")

		assert.Equal(SYMLINK, link_to_file.Type())
		assert.Equal(link_to_file, folder.Get("link_to_file"))
		assert.Equal("target_file", link_to_file.Target())

		assert.Equal(SYMLINK, link_to_folder.Type())
		assert.Equal(link_to_folder, folder.Get("link_to_folder"))
		assert.Equal("target_folder", link_to_folder.Target())
	})

	t.Run("error cases", func(t *testing.T) {
		assert := assert.New(t)
		folder := folder.Clone().(*Folder)

		broken_link := folder.Symlink("broken_link", "nowhere")
		assert.Equal(broken_link, folder.Get("broken_link"))
		assert.Equal(SYMLINK, broken_link.Type())

		outside_link := folder.Symlink("outside_link", "../somewhere-else")
		assert.Equal(outside_link, folder.Get("outside_link"))
		assert.Equal(SYMLINK, outside_link.Type())

		// Test cyclic reference
		cyclic_link := folder.Symlink("cyclic_link", ".")
		assert.Equal(cyclic_link, folder.Get("cyclic_link"))
		assert.Equal(SYMLINK, cyclic_link.Type())
	})

	// TODO: implement
	// t.Run("persistence", func(t *testing.T) {
	// 	assert := assert.New(t)
	// 	folder := folder.Clone().(*Folder)
	//
	// 	// Write to disk
	// 	tempDir := t.TempDir()
	// 	assert.NoError(folder.WriteTo(path.Join(tempDir, "folder")))
	//
	// 	// Read back from disk
	// 	loadedFolder := NewFolder()
	// 	assert.NoError(loadedFolder.ReadFrom(path.Join(tempDir, "folder")))
	//
	// 	// Verify links are preserved
	// 	link_to_file := loadedFolder.Get("link_to_file").(*Symlink)
	//
	// 	assert.Equal(SYMLINK, link_to_file.Type())
	// 	assert.Equal("target_file", link_to_file.Target)
	//
	// 	// Test diffing
	// 	diff := folder.Diff(loadedFolder)
	// 	assert.Empty(diff, "Folders should be identical after load")
	// })
}

func TestHardlink(t *testing.T) {
	folder := NewFolder()

	// Create some files to link to
	folder.FileString("original_file", "hello world")

	t.Run("basic hardlink operations", func(t *testing.T) {
		assert := assert.New(t)

		folder := folder.Clone().(*Folder)
		// Create valid hardlink
		hard_link := folder.Hardlink("hard_link", "original_file")

		// Verify the link exists and is correct type
		assert.Equal(hard_link, folder.Get("hard_link"))
		assert.Equal(HARDLINK, hard_link.Type())
		assert.Equal("original_file", hard_link.Target())

		// Verify content is accessible through both paths
		original_file := folder.Get("original_file").(*File)
		assert.Equal("hello world", original_file.ContentString())

		// TODO: should we support this?
		// assert.Equal("hello world", hard_link.GetContent())
	})

	t.Run("error cases", func(t *testing.T) {
		// assert := assert.New(t)

		folder := folder.Clone().(*Folder)
		broken_link := folder.Hardlink("broken_link", "nowhere")
		some_dir := folder.Folder("some_dir")
		dir_link := folder.Hardlink("dir_link", "some_dir")
		outside_link := folder.Hardlink("outside_link", "../somewhere-else")

		_, _, _, _ = broken_link, some_dir, dir_link, outside_link

		// TODO: more
	})

	// TODO:  do this
	// t.Run("persistence", func(t *testing.T) {
	// 	assert := assert.New(t)
	//
	// 	folder := folder.Clone().(*Folder)
	// 	// Write to disk
	// 	tempDir := t.TempDir()
	// 	assert.NoError(folder.WriteTo(tempDir))
	//
	// 	// Read back from disk
	// 	loadedFolder := NewFolder()
	// 	assert.NoError(loadedFolder.ReadFrom(tempDir))
	//
	// 	// Verify hardlinks are preserved
	// 	hard_link := loadedFolder.Get("hard_link").(*HardLink)
	// 	assert.Equal(HARDLINK, hard_link.Type())
	// 	assert.Equal("original_file", hard_link.Target)
	//
	// 	// Test diffing
	// 	diff := folder.Diff(loadedFolder)
	// 	assert.Empty(diff, "Folders should be identical after load")
	//
	// 	// TODO: more
	// })
}
