package fs

import (
	"sort"
	"strings"
)

type Operand string

const (
	Unlink       Operand = "Unlink"
	Rmdir        Operand = "Rmdir"
	Mkdir        Operand = "Mkdir"
	Create       Operand = "Create"
	ChangeFile   Operand = "ChangeFile"
	ChangeFolder Operand = "ChangeDir"
)

type Reason struct {
	Type   ReasonType
	Before interface{}
	After  interface{}
}

// TODDO: expand reason from enum, to struct (before / after)
type ReasonType string

const (
	TypeChanged    ReasonType = "Type Changed"
	ModeChanged    ReasonType = "Mode Changed"
	ContentChanged ReasonType = "Content Changed"
	Missing        ReasonType = "Missing"
	Because        ReasonType = "because"
)

type Operation struct {
	RelativePath string
	Operand      Operand
	Reason       Reason
	Operations   []Operation
}

// Assume: a and b are the same root, compare a to b, and provide the patch
// required to transform A to B, using the same efficient protocol as node fs-tree-diff
//
// unlink: remove a file
// rmdir: remove a directory
// mkdir: create a directory
// create: create a file
// change: change
//
// given this is go, it could be interesting to eventually make this streaming,
// so we can caluclate while starting computation based on the partial result.
// In many cases,this could allow the CPU to start building, even though we are
// still calcuating. Food for thought.
func sortStringsToLower(slice []string) {
	sort.Slice(slice, func(i, j int) bool {
		return strings.ToLower(slice[i]) < strings.ToLower(slice[j])
	})
}

func Diff(a, b *Folder, caseSensitive bool) []Operation {
	updates := []Operation{}
	additions := []Operation{}
	removals := []Operation{}

	a_index := 0
	b_index := 0
	a_keys := a.Entries()
	b_keys := b.Entries()

	if !caseSensitive {
		sortStringsToLower(a_keys)
		sortStringsToLower(b_keys)
	}

	// iterate over both arrays subset of those arrays that are the same length
	for a_index < len(a_keys) && b_index < len(b_keys) {
		a_key := a_keys[a_index]
		b_key := b_keys[b_index]

		a_comparable := a_key
		b_comparable := b_key

		if !caseSensitive {
			a_comparable = strings.ToLower(a_comparable)
			b_comparable = strings.ToLower(b_comparable)
		}

		if a_comparable == b_comparable {
			a_entry := a.Get(a_key)
			b_entry := b.Get(b_key)

			a_type := a_entry.Type()
			b_type := b_entry.Type()

			if a_type != b_type {
				// if they are not the same type, easy
				// we remove a
				removals = append(removals, a.RemoveChildOperation(a_key))
				// and then we add b
				additions = append(additions, b.CreateChildOperation(b_key))
			} else if a_type == FOLDER {
				// they are both folders, so we recurse
				operations := Diff(a_entry.(*Folder), b_entry.(*Folder), caseSensitive)

				if len(operations) > 0 {
					updates = append(updates, NewChangeFolder(a_key, operations...))
				}
			} else if a_type == FILE {
				equal, reason := a_entry.EqualWithReason(b_entry)
				if equal {
					// they are equal files, so do nothing..
				} else {
					// they are different then so we add a change operation
					updates = append(updates, NewChangeFile(b_key, reason))
				}
			} else {
				panic("fsdt/diff.go(unreachable)")
			}

			a_index++
			b_index++
		} else if a_key < b_key {
			// a is missing from b
			a_index++
			removals = append(removals, a.RemoveChildOperation(a_key))
		} else if a_key > b_key {
			// b is missing form a
			b_index++
			removals = append(removals, b.CreateChildOperation(b_key))
		} else {
			panic("fsdt/diff.go(unreachable)")
		}
	}

	// either both, or one of the arrays is exhausted
	// if stuff remains in A, remove them
	for a_index < len(a_keys) {
		relative_path := a_keys[a_index]
		a_index++
		removals = append(removals, a.RemoveChildOperation(relative_path))
	}

	// if stuff remains in B, add them
	for b_index < len(b_keys) {
		relative_path := b_keys[b_index]
		b_index++
		additions = append(additions, b.CreateChildOperation(relative_path))
	}

	return append(append(removals, updates...), additions...)
}

func handleVariadicOperation(operation Operand, relativePath string, operations ...Operation) Operation {
	if len(operations) == 0 {
		return Operation{Operand: operation, RelativePath: relativePath, Operations: nil}
	}
	return Operation{Operand: operation, RelativePath: relativePath, Operations: operations}
}

func NewRmdir(relativePath string, operations ...Operation) Operation {
	return handleVariadicOperation(Rmdir, relativePath, operations...)
}

func NewMkdir(relativePath string, operations ...Operation) Operation {
	return handleVariadicOperation(Mkdir, relativePath, operations...)
}

func NewUnlink(relativePath string) Operation {
	return Operation{Operand: Unlink, RelativePath: relativePath}
}

func NewCreate(relativePath string) Operation {
	return Operation{Operand: Create, RelativePath: relativePath}
}

func NewChangeFile(relativePath string, reason Reason) Operation {
	return Operation{Operand: ChangeFile, RelativePath: relativePath, Reason: reason}
}

func NewChangeFolder(relativePath string, operations ...Operation) Operation {
	return Operation{Operand: ChangeFolder, RelativePath: relativePath, Operations: operations}
}
