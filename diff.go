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

// DiffOptions configures how the diff operation should behave
type DiffOptions struct {
	CaseSensitive     bool
	CompareMode       bool
	ComparePermissions bool
	CompareTimestamps  bool
	CompareContent     bool
	CompareAttributes  bool
	IncludeFolderName  bool
}

// DefaultDiffOptions returns sensible default diff options
func DefaultDiffOptions() DiffOptions {
	return DiffOptions{
		CaseSensitive:     true,
		CompareMode:       true,
		ComparePermissions: true,
		CompareTimestamps:  false,
		CompareContent:     true,
		CompareAttributes:  false,
		IncludeFolderName:  true,
	}
}

// Diff compares two folders and returns the operations needed to transform a into b
func Diff(a, b *Folder, options DiffOptions) op.Operation {
	dirValue := op.DirectoryValue{}

	a_index := 0
	b_index := 0
	a_keys := a.Entries()
	b_keys := b.Entries()

	if !options.CaseSensitive {
		sortStringsToLower(a_keys)
		sortStringsToLower(b_keys)
	}

	// iterate over both arrays subset of those arrays that are the same length
	for a_index < len(a_keys) && b_index < len(b_keys) {
		a_key := a_keys[a_index]
		b_key := b_keys[b_index]

		a_comparable := a_key
		b_comparable := b_key

		if !options.CaseSensitive {
			a_comparable = strings.ToLower(a_comparable)
			b_comparable = strings.ToLower(b_comparable)
		}

		if a_comparable == b_comparable {
			a_entry := a.Get(a_key)
			b_entry := b.Get(b_key)

			a_type := a_entry.Type()
			b_type := b_entry.Type()

			equal, reason := compareEntries(a_entry, b_entry, options)

			if equal {
				// do nothing!
			} else if a_type == FOLDER && b_type == FOLDER {
				a_entry := a_entry.(*Folder)
				b_entry := b_entry.(*Folder)

				// they are both folders, so we recurse
				operation := Diff(a_entry, b_entry, options)

				// Set the relative path based on options
				if options.IncludeFolderName {
					operation.RelativePath = b_key
				} else {
					operation.RelativePath = "."
				}

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
			// This should never happen, but handle gracefully
			if a_key < b_key {
				a_index++
				dirValue.AddOperations(a.RemoveChildOperation(a_key, op.Reason{}))
			} else {
				b_index++
				dirValue.AddOperations(b.CreateChildOperation(b_key, op.Reason{}))
			}
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

	// Set the relative path based on options
	relativePath := "."
	if options.IncludeFolderName {
		relativePath = "."
	}

	return op.NewChangeDirOperation(relativePath, dirValue.Operations...)
}

// compareEntries compares two entries with the given options and returns equality and reason
func compareEntries(a, b FolderEntry, options DiffOptions) (bool, op.Reason) {
	// Type comparison is always done
	if a.Type() != b.Type() {
		return false, op.Reason{
			Type:    op.ReasonTypeChanged,
			Before:  a.Type(),
			After:   b.Type(),
			Message: "Entry type changed",
		}
	}

	// Mode comparison - only if both entries support it
	if options.CompareMode {
		// Check if both entries have mode information
		if aFile, aOk := a.(*File); aOk {
			if bFile, bOk := b.(*File); bOk {
				if aFile.Mode() != bFile.Mode() {
					return false, op.Reason{
						Type:    op.ReasonModeChanged,
						Before:  aFile.Mode(),
						After:   bFile.Mode(),
						Message: "File mode changed",
					}
				}
			}
		}
		
		if aFolder, aOk := a.(*Folder); aOk {
			if bFolder, bOk := b.(*Folder); bOk {
				if aFolder.Mode() != bFolder.Mode() {
					return false, op.Reason{
						Type:    op.ReasonModeChanged,
						Before:  aFolder.Mode(),
						After:   bFolder.Mode(),
						Message: "Folder mode changed",
					}
				}
			}
		}
		
		if aLink, aOk := a.(*Link); aOk {
			if bLink, bOk := b.(*Link); bOk {
				if aLink.Mode() != bLink.Mode() {
					return false, op.Reason{
						Type:    op.ReasonModeChanged,
						Before:  aLink.Mode(),
						After:   bLink.Mode(),
						Message: "Link mode changed",
					}
				}
			}
		}
	}

	// Content comparison
	if options.CompareContent {
		equal, reason := a.EqualWithReason(b)
		if !equal {
			return false, reason
		}
	}

	return true, op.Reason{}
}

// Legacy Diff function for backward compatibility
func DiffLegacy(a, b *Folder, caseSensitive bool) op.Operation {
	options := DefaultDiffOptions()
	options.CaseSensitive = caseSensitive
	return Diff(a, b, options)
}
