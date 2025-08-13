package fsdt

import (
	"testing"

	op "github.com/stefanpenner/go-fsdt/operation"
	"github.com/stretchr/testify/assert"
)

func TestDiffStuffEmpty(t *testing.T) {
	assert := assert.New(t)

	a := NewFolder()
	b := NewFolder()

	assert.True(a.Diff(b).IsNoop())
}

func TestDiffWithDifferentCase(t *testing.T) {
	assert := assert.New(t)
	// basically, we are case sensitive for now.
	// TODO: handle case insensitivity, including Apples default approach.
	// were fires can be store case sensitively, but are resolved case
	// insensitively. A file can be README.md or readme.md, but the file
	// system will only store one of them, and allow a uniform access
	// to the file.
	//
	// To handle this we must:
	// * Most likely we must sort by the lower case version of the file name
	// * Access via the case used
	// * Compare via the lower case
	// * Provide the operations based on the intended case
	//
	// But this can be for another day
	//
	a := NewFolder()
	b := NewFolder()

	a.FileString("README.md", "## HI\n")
	a.FileString("a.md", "## HI\n")
	a.FileString("keep.md", "## HI\n")
	a.Symlink("b.md", "a.md")

	b.FileString("readme.md", "## HI\n")
	b.FileString("b.md", "## HI\n")
	b.FileString("keep.md", "## HI\n")
	b.Symlink("B.md", "b.md")

	assert.Equal(
		op.NewChangeDirOperation(".",
			// case sensitive
			op.NewCreateLinkOperation("B.md", "b.md", op.SymbolicLink, 0777),
			op.NewRemoveFileOperation("README.md"),
			op.NewRemoveFileOperation("a.md"),
			op.NewRemoveLinkOperation("b.md"),
			op.NewCreateFileOperation("b.md", []byte("## HI\n"), 0644),
			op.NewCreateFileOperation("readme.md", []byte("## HI\n"), 0644),
		), a.Diff(b))

	assert.Equal(op.NewChangeDirOperation(".",
		op.NewCreateLinkOperation("B.md", "b.md", op.SymbolicLink, 0777),
		op.NewRemoveFileOperation("a.md"),
		op.NewRemoveLinkOperation("b.md"),
		op.NewCreateFileOperation("b.md", []byte("## HI\n"), 0644),
	), a.CaseInsensitiveDiff(b))
}

func TestDiffStuffAWithEmptyB(t *testing.T) {
	assert := assert.New(t)

	a := NewFolder()
	b := NewFolder()

	a.FileString("README.md", "## HI\n")
	a.FileString("BUILD.bazel", "## HI\n")
	a.Folder("apple", func(f *Folder) {})
	a.Folder("lib", func(f *Folder) {})
	a.Folder("apple", func(f *Folder) {})
	a.Symlink("a", "apple")

	assert.Equal(op.NewChangeDirOperation(".",
		op.NewRemoveFileOperation("BUILD.bazel"),
		op.NewRemoveFileOperation("README.md"),
		op.NewRemoveLinkOperation("a"),
		op.NewRemoveDirOperation("apple"),
		op.NewRemoveDirOperation("lib"),
	), a.Diff(b))
}

func TestDiffStuffBWithEmptyA(t *testing.T) {
	assert := assert.New(t)

	a := NewFolder()
	b := NewFolder()

	b.FileString("README.md", "## HI\n")
	b.FileString("BUILD.bazel", "## HI\n")
	b.Folder("lib")
	b.Folder("apple")
	b.Symlink("a", "apple")

	assert.Equal(op.NewChangeDirOperation(".",
		op.NewCreateFileOperation("BUILD.bazel", []byte("## HI\n"), 0644),
		op.NewCreateFileOperation("README.md", []byte("## HI\n"), 0644),
		op.NewCreateLinkOperation("a", "apple", op.SymbolicLink, 0777),
		op.NewCreateDirOperation("apple"),
		op.NewCreateDirOperation("lib"),
	), a.Diff(b))
}

func TestDiffStuffWithOverlap(t *testing.T) {
	assert := assert.New(t)

	a := NewFolder()
	b := NewFolder()

	a.FileString("README.md", "## HI\n")
	b.FileString("README.md", "## HI\n")

	a.Symlink("a", "somewhere")
	b.Symlink("a", "somewhere")

	a.Symlink("d", "somewhere")
	b.Symlink("d", "somewhere-else")

	a.FileString("BUILD.bazel", "## HI\n")
	// BUILD.bazel is not in b

	a.Folder("apple")
	b.Folder("apple")

	// orange is not in a
	b.Folder("orange")

	// NOTES.txt is not in a
	b.FileString("notes.txt", "## HI\n")

	a.Folder("lib")
	// lib is not in b

	assert.Equal(op.NewChangeDirOperation(".",
		op.NewRemoveFileOperation("BUILD.bazel"),
		op.NewRemoveLinkOperation("d"),
		op.NewCreateLinkOperation("d", "somewhere-else", op.SymbolicLink, 0777),
		op.NewRemoveDirOperation("lib"),
		op.NewCreateFileOperation("notes.txt", []byte("## HI\n"), 0644),
		op.NewCreateDirOperation("orange"),
	), a.Diff(b))
}

func TestWithContentDifferences(t *testing.T) {
	assert := assert.New(t)

	a := NewFolder()
	b := NewFolder()

	readme := a.FileString("README.md", "## HI\n")
	b.FileString("README.md", "## Bye\n")

	assert.Equal(op.NewChangeDirOperation(".",
		readme.ChangeOperation("README.md", op.Reason{
			Type:   op.ReasonContentChanged,
			Before: []byte("## HI\n"),
			After:  []byte("## Bye\n"),
		}),
	), a.Diff(b))
}

func TestDiffWithDepth(t *testing.T) {
	assert := assert.New(t)

	a := NewFolder()
	b := NewFolder()

	a.Folder("foo", func(f *Folder) {
		f.FileString("README.md", "## HI\n")
		f.Folder("bar", func(f *Folder) {
			f.FileString("a.md", "## HI\n")
			f.FileString("README.md", "## HI\n")
		})
	})

	b.Folder("foo", func(f *Folder) {
		f.FileString("_README.md", "## BYE\n")
		f.Folder("_bar", func(f *Folder) {
			f.FileString("_b.md", "## HI\n")
			f.FileString("_README.md", "## BYE\n")
		})
	})

	assert.Equal(op.NewChangeDirOperation(".",
		op.NewChangeDirOperation("foo",
			op.NewRemoveFileOperation("README.md"),
			op.NewCreateFileOperation("_README.md", []byte("## BYE\n"), 0644),
			op.NewCreateDirOperation("_bar",
				op.NewCreateFileOperation("_README.md", []byte("## BYE\n"), 0644),
				op.NewCreateFileOperation("_b.md", []byte("## HI\n"), 0644),
			),
			op.NewRemoveDirOperation("bar",
				op.NewRemoveFileOperation("README.md"),
				op.NewRemoveFileOperation("a.md"),
			),
		),
	), a.Diff(b))
}

func TestDiffWithDepthAndContent(t *testing.T) {
	assert := assert.New(t)

	a := NewFolder()
	b := NewFolder()

	a.Folder("foo", func(f *Folder) {
		f.Folder("bar", func(f *Folder) {
			f.FileString("a.md", "## HI\n")
			f.FileString("README.md", "## HI\n")
		})
	})
	readme := a.FileString("README.md", "## HI\n")

	b.Folder("foo", func(f *Folder) {
		f.FileString("README.md", "## BYE\n")
		f.Folder("bar", func(f *Folder) {
			f.FileString("b.md", "## HI\n")
			f.FileString("README.md", "## BYE\n")
		})
	})

	assert.Equal(
		readme.ChangeOperation("apple", op.Reason{}).String(),
		readme.ChangeOperation("apple", op.Reason{}).String(),
	)

	expected := op.NewChangeDirOperation(".",
		readme.RemoveOperation("README.md", op.Reason{}),
		op.NewChangeDirOperation("foo",
			op.NewCreateFileOperation("README.md", []byte("## BYE\n"), 0644),
			op.NewChangeDirOperation("bar",

				readme.ChangeOperation("README.md", op.Reason{
					Type:   op.ReasonContentChanged,
					Before: []byte("## HI\n"),
					After:  []byte("## BYE\n"),
				}),
				op.NewRemoveFileOperation("a.md"),
				op.NewCreateFileOperation("b.md", []byte("## HI\n"), 0644),
			),
		),
	)

	assert.Equal(
		expected.String(),
		a.Diff(b).String(),
	)
}
