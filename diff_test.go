package fsdt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffStuffEmpty(t *testing.T) {
	assert := assert.New(t)

	a := NewFolder()
	b := NewFolder()
	Nothing := []Operation{}

	assert.Equal(Nothing, a.Diff(b))
}

func TestDiffWithDifferentCase(t *testing.T) {
	assert := assert.New(t)
	// basically, we are case sensitive for now.
	// TODO: handle case insensitivity, including Apples default approach.
	// were files can be store case sensitively, but are resolved case
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

	b.FileString("readme.md", "## HI\n")
	b.FileString("b.md", "## HI\n")
	b.FileString("keep.md", "## HI\n")

	// case sensitive
	assert.Equal([]Operation{
		NewUnlink("README.md"),
		NewUnlink("a.md"),
		NewCreate("b.md"),
		NewCreate("readme.md"),
	}, a.Diff(b))

	assert.Equal([]Operation{
		NewUnlink("a.md"),
		NewCreate("b.md"),
	}, a.CaseInsensitiveDiff(b))
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

	assert.Equal([]Operation{
		NewUnlink("BUILD.bazel"),
		NewUnlink("README.md"),
		NewRmdir("apple"),
		NewRmdir("lib"),
	}, a.Diff(b))
}

func TestDiffStuffBWithEmptyA(t *testing.T) {
	assert := assert.New(t)

	a := NewFolder()
	b := NewFolder()

	b.FileString("README.md", "## HI\n")
	b.FileString("BUILD.bazel", "## HI\n")
	b.Folder("lib")
	b.Folder("apple")

	assert.Equal([]Operation{
		NewCreate("BUILD.bazel"),
		NewCreate("README.md"),
		NewMkdir("apple"),
		NewMkdir("lib"),
	}, a.Diff(b))
}

func TestDiffStuffWithOverlap(t *testing.T) {
	assert := assert.New(t)

	a := NewFolder()
	b := NewFolder()

	a.FileString("README.md", "## HI\n")
	b.FileString("README.md", "## HI\n")

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

	assert.Equal([]Operation{
		NewUnlink("BUILD.bazel"),
		NewRmdir("lib"),
		NewCreate("notes.txt"),
		NewMkdir("orange"),
	}, a.Diff(b))
}

func TestWithContentDifferences(t *testing.T) {
	assert := assert.New(t)

	a := NewFolder()
	b := NewFolder()

	a.FileString("README.md", "## HI\n")
	b.FileString("README.md", "## Bye\n")

	assert.Equal([]Operation{
		NewChangeFile("README.md", Reason{
			Type:   ContentChanged,
			Before: []byte("## HI\n"),
			After:  []byte("## Bye\n"),
		}),
	}, a.Diff(b))
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

	assert.Equal([]Operation{
		NewChangeFolder("foo",
			NewUnlink("README.md"),
			NewCreate("_README.md"),
			NewMkdir("_bar", NewCreate("_README.md"), NewCreate("_b.md")),
			NewRmdir("bar", NewUnlink("README.md"), NewUnlink("a.md")),
		),
	}, a.Diff(b))
}

func TestDiffWithDepthAndContent(t *testing.T) {
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
		f.FileString("README.md", "## BYE\n")
		f.Folder("bar", func(f *Folder) {
			f.FileString("b.md", "## HI\n")
			f.FileString("README.md", "## BYE\n")
		})
	})

	assert.Equal([]Operation{
		NewChangeFolder("foo",
			NewChangeFile("README.md", Reason{
				Type:   ContentChanged,
				Before: []byte("## HI\n"),
				After:  []byte("## BYE\n"),
			}),
			NewChangeFolder("bar",
				NewUnlink("a.md"),
				NewChangeFile("README.md", Reason{
					Type:   ContentChanged,
					Before: []byte("## HI\n"),
					After:  []byte("## BYE\n"),
				}),
				NewCreate("b.md"),
			),
		),
	}, a.Diff(b))
}
