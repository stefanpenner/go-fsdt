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

	assert.Equal(op.Nothing, a.Diff(b))
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
		op.NewChangeFolderOperation(".",
			// case sensitive
			op.NewCreateLink("B.md", "b.md"),
			op.NewUnlink("README.md"),
			op.NewUnlink("a.md"),
			op.NewUnlink("b.md"),
			op.NewFileOperation("b.md"),
			op.NewFileOperation("readme.md"),
		), a.Diff(b))

	assert.Equal(op.NewChangeFolderOperation(".",
		op.NewCreateLink("B.md", "b.md"),
		op.NewUnlink("a.md"),
		op.NewUnlink("b.md"),
		op.NewFileOperation("b.md"),
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

	assert.Equal(op.NewChangeFolderOperation(".",
		op.NewUnlink("BUILD.bazel"),
		op.NewUnlink("README.md"),
		op.NewUnlink("a"),
		op.NewRmdir("apple"),
		op.NewRmdir("lib"),
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

	assert.Equal(op.NewChangeFolderOperation(".",
		op.NewFileOperation("BUILD.bazel"),
		op.NewFileOperation("README.md"),
		op.NewCreateLink("a", "apple"),
		op.NewMkdirOperation("apple"),
		op.NewMkdirOperation("lib"),
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

	assert.Equal(op.NewChangeFolderOperation(".",
		op.NewUnlink("BUILD.bazel"),
		op.NewUnlink("d"),
		op.NewCreateLink("d", "somewhere-else"),
		op.NewRmdir("lib"),
		op.NewFileOperation("notes.txt"),
		op.NewMkdirOperation("orange"),
	), a.Diff(b))
}

func TestWithContentDifferences(t *testing.T) {
	assert := assert.New(t)

	a := NewFolder()
	b := NewFolder()

	readme := a.FileString("README.md", "## HI\n")
	b.FileString("README.md", "## Bye\n")

	assert.Equal(op.NewChangeFolderOperation(".",
		readme.ChangeOperation("README.md", op.Reason{
			Type:   op.ContentChanged,
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

	assert.Equal(op.NewChangeFolderOperation(".",
		op.NewChangeFolderOperation("foo",
			op.NewUnlink("README.md"),
			op.NewFileOperation("_README.md"),
			op.NewMkdirOperation("_bar",
				op.NewFileOperation("_README.md"),
				op.NewFileOperation("_b.md"),
			),
			op.NewRmdir("bar",
				op.NewUnlink("README.md"),
				op.NewUnlink("a.md"),
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
		op.Print(readme.ChangeOperation("apple", op.Reason{})),
		op.Print(readme.ChangeOperation("apple", op.Reason{})),
	)

	expected := op.Print(
		op.NewChangeFolderOperation(".",
			readme.RemoveOperation("README.md", op.Reason{}),
			op.NewChangeFolderOperation("foo",
				op.NewFileOperation("README.md"),
				op.NewChangeFolderOperation("bar",

					readme.ChangeOperation("README.md", op.Reason{
						Type:   op.ContentChanged,
						Before: []byte("## HI\n"),
						After:  []byte("## BYE\n"),
					}),
					op.NewUnlink("a.md"),
					op.NewFileOperation("b.md"),
				),
			),
		),
	)

	assert.Equal(
		expected,
		op.Print(a.Diff(b)),
	)
}
