package fsdt

import (
	"testing"

	op "github.com/stefanpenner/go-fsdt/operation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSymlink(t *testing.T) {
	folder := NewFolder()

	folder.File("target_file", FileOptions{Content: []byte("hello")})
	folder.Folder("target_folder")
	folder.Symlink("link_to_file", "target_file")
	folder.Symlink("link_to_folder", "target_folder")

	t.Run("basic symlink operations", func(t *testing.T) {
		assert := assert.New(t)
		folder := folder.Clone().(*Folder)
		link_to_file := folder.Get("link_to_file").(*Link)
		link_to_folder := folder.Get("link_to_folder").(*Link)

		assert.Equal(SYMLINK, link_to_file.Type())
		assert.Equal(link_to_file, folder.Get("link_to_file"))
		assert.Equal("target_file", link_to_file.Target())

		assert.Equal(SYMLINK, link_to_folder.Type())
		assert.Equal(link_to_folder, folder.Get("link_to_folder"))
		assert.Equal("target_folder", link_to_folder.Target())

		broken_link := folder.Symlink("broken_link", "nowhere")
		assert.Equal(broken_link, folder.Get("broken_link"))
		assert.Equal(SYMLINK, broken_link.Type())

		outside_link := folder.Symlink("outside_link", "../somewhere-else")
		assert.Equal(outside_link, folder.Get("outside_link"))
		assert.Equal(SYMLINK, outside_link.Type())

		cyclic_link := folder.Symlink("cyclic_link", ".")
		assert.Equal(cyclic_link, folder.Get("cyclic_link"))
		assert.Equal(SYMLINK, cyclic_link.Type())
	})

	t.Run("Equal", func(t *testing.T) {
		assert := assert.New(t)
		folder := folder.Clone().(*Folder)

		symlink := folder.Symlink("from", "to")
		other := folder.Symlink("from", "to")
		a := folder.Symlink("from", "not-to")
		b := folder.Symlink("not-from", "from")
		c := folder.Symlink("other", "from")

		assert.True(symlink.Equal(symlink))
		assert.True(symlink.Equal(symlink.Clone()))
		assert.True(symlink.Equal(other))
		assert.False(symlink.Equal(a))
		assert.False(symlink.Equal(b))
		assert.True(b.Equal(c))
	})

	t.Run("EqualWithReason", func(t *testing.T) {
		t.Run("symlink", func(t *testing.T) {
			assert := assert.New(t)
			folder := folder.Clone().(*Folder)

			symlink := folder.Symlink("from", "to")
			other := folder.Symlink("from", "to")
			a := folder.Symlink("from", "not-to")
			b := folder.Symlink("not-from", "from")
			c := folder.Symlink("other", "from")

			equal, reason := symlink.EqualWithReason(symlink.Clone())
			assert.True(equal)
			assert.Equal(op.Reason{}, reason)

			equal, reason = symlink.EqualWithReason(other)
			assert.True(equal)
			assert.Equal(op.Reason{}, reason)

			equal, reason = symlink.EqualWithReason(a)
			assert.False(equal)
			assert.Equal(op.Reason{Type: op.ReasonContentChanged, Before: "to", After: "not-to"}, reason)

			equal, reason = symlink.EqualWithReason(b)
			assert.False(equal)
			assert.Equal(op.Reason{Type: op.ReasonContentChanged, Before: "to", After: "from"}, reason)

			equal, reason = b.EqualWithReason(c)
			assert.True(equal)
			assert.Equal(op.Reason{}, reason)
		})
	})

	t.Run("persistence", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)
		folder := folder.Clone().(*Folder)

		// Write to disk
		location := t.TempDir()
		require.NoError(folder.WriteTo(location))

		// Read back from disk
		loadedFolder := NewFolder()
		require.NoError(loadedFolder.ReadFrom(location))

		// Verify links are preserved
		link_to_file := loadedFolder.Get("link_to_file").(*Link)

		assert.Equal(SYMLINK, link_to_file.Type())
		assert.Equal("target_file", link_to_file.Target())

		// Test diffing
		diff := folder.Diff(loadedFolder)
		assert.True(diff.IsNoop(), "Folders should be identical after load")
	})
}
