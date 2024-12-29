# go-fsdt

**A File System Diffing Toolchain**

> ⚠️ **Alpha Status**: This project is in early development. APIs may change without notice.

`go-fsdt` is a reimagined Go implementation of several Node.js libraries,
offering powerful tools for comparing file systems, deriving minimal IO
operations, and managing file system fixtures.

#### But why?
In my experience, eventually these capabilities all need to work together, so
this time, I'm building them together from the start.

Inspired by:
- [fs-tree-diff](https://github.com/stefanpenner/fs-tree-diff)
- [node-fixturify-project](https://github.com/stefanpenner/node-fixturify-project)
- [node-fixturify](https://github.com/joliss/node-fixturify)

---

## Features

- Compare two directories and identify differences.
- Generate minimal IO operations to transform one directory tree into another.
- Create and manage file system fixtures for testing.
- Seamlessly compare file systems between test runs.

---

## When to Use

`go-fsdt` is ideal for:

- Testing and validation of file system changes.
- Optimizing file system operations for performance.
- Creating or comparing directory structures in test environments.

---

## Examples


Basic Example
```go
// create a new folder, in memory
folder := fsdt.NewFolder(func(f *fsdt.Folder) {
  f.FileString("todo.txt", "- [] finish FSDT\n- [] ...\n- [ ] Profit")
})

// add a file, with the content of a string
readme_one := folder.FileString("README_one.md", "## h1\n")

// add a file, with the content of a []byte{} and other more advanced options
readme_two := folder.File("README_two.md", fsdt.FileOptions {
  Content: []byte("## h1 \n")
})

// create lib/Empty.txt
lib := folder.Folder("lib", func(lib *fsdt.Folder) {
  // create another file, this time with no content
  lib.File("Empty.txt")
})

// empty folder
tests := folder.Folder("tests")

// write that folder to disk
err := folder.WriteTo("/path/to/somewhere")

// now /path/to/somewhere has the contents of folder

// let's create a new folder from disk
newFolder, error := fsdt.ReadFrom("/path/to/somewhere")
require.NoError(err)

// let's compare the two, but they are the same so it's boring
newFolder.Diff(folder) // => []

// we can also create a new folder via clone
clone := folder.Clone().(*Folder)

// let's compare the two, but they are the same so it's boring
folder.Diff(folder) // => []
new_folder.Diff(folder) // => []

// let's remove a folder and file
err = folder.Remove("lib")

// now there should be a diff, notice the diff is structure for efficient disk
// transformations.
assert.Equal(op.Operation{
  RelativePath: "lib",
  Operand: Mkdir,
  Operations: []Operation {
    {
      RelativePath: "Empty.txt",
      Operand: Create
    }
  }
}, folder.Diff(clone.(*Folder)))

// and this should be the reverse.
assert.Equal(op.Operation{
  RelativePath: "lib",
  Operand:      Rmdir,
  Operations: []Operation{
    {
      RelativePath: "Empty.txt",
      Operand:      Unlink,
    },
  },
}, clone.(*Folder).Diff(folder))
```

---

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

---

## License

[MIT License](LICENSE)
