package fsdt

import (
	"sort"
	"strings"

	op "github.com/stefanpenner/go-fsdt/operation"
)

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

func Diff(a, b *Folder, caseSensitive bool) op.Operation {
	dirValue := op.DirectoryValue{}

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

			equal, reason := a_entry.EqualWithReason(b_entry)

			if equal {
				// do nothing!
			} else if a_type == FOLDER && b_type == FOLDER {
				a_entry := a_entry.(*Folder)
				b_entry := b_entry.(*Folder)

				// TODO: folder modes, permissions etc. can change
				// they are both folders, so we recurse
				operation := Diff(a_entry, b_entry, caseSensitive)

				operation.RelativePath = b_key
				if !operation.IsNoop() {
					dirValue.AddOperations(operation)
				}
			} else if a_type == FILE && b_type == FILE {
				dirValue.AddOperations(a_entry.ChangeOperation(b_key, reason))
			} else {
				dirValue.AddOperations(
					a_entry.RemoveOperation(b_key, reason),
					b_entry.CreateOperation(b_key, reason),
				)
			}

			a_index++
			b_index++
		} else if a_key < b_key {
			// a is missing from b
			a_index++

			dirValue.AddOperations(a.RemoveChildOperation(a_key, op.Reason{})) // TODO: missing reason
		} else if b_key < a_key {
			// b is missing form a
			b_index++
			dirValue.AddOperations(b.CreateChildOperation(b_key, op.Reason{})) // TODO: missing reason
		} else {
			panic("fsdt/diff.go(unreachable)")
		}
	}

	// either both, or one of the arrays is exhausted
	// if stuff remains in A, remove them
	for a_index < len(a_keys) {
		dirValue.AddOperations(a.RemoveChildOperation(a_keys[a_index], op.Reason{})) // TODO: missing reason
		a_index++
	}

	// if stuff remains in B, create them
	for b_index < len(b_keys) {
		dirValue.AddOperations(b.CreateChildOperation(b_keys[b_index], op.Reason{})) // TODO: missing reason
		b_index++
	}

	if len(dirValue.Operations) == 0 {
		return op.NewNoopOperation()
	}

	return op.NewChangeDirOperation(".", dirValue.Operations...)
}
