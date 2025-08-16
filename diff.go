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

// FileContentStrategy controls how file content equality is determined.
type FileContentStrategy int

const (
	// Compare file bytes directly
	CompareBytes FileContentStrategy = iota
	// Prefer checksum/xattr if present; fall back to bytes when unavailable or algorithms mismatch
	PreferChecksumOrBytes
	// Require checksum-only comparison; if absent for either side, treat as changed
	RequireChecksum
	// Skip comparing file content (only mode/type/name differences will be detected)
	SkipContent
)

// DiffOptions tunes diff behavior for performance vs. thoroughness.
type DiffOptions struct {
	CaseSensitive bool
	ContentStrategy FileContentStrategy
	// Optional: when computing or comparing checksums, the expected algorithm name, e.g. "sha256"
	ChecksumAlgorithm string
	// Optional metadata comparisons
	CompareMode  bool // default true
	CompareSize  bool // default false
	CompareMTime bool // default false
	// If true, compute checksum when missing and ContentStrategy needs one (for in-memory trees)
	ComputeChecksumIfMissing bool
	// If true, when ComputeChecksumIfMissing occurs and a file has a source path, write checksum to xattr
	WriteComputedChecksumToXAttr bool
	// If set, xattr key to write when WriteComputedChecksumToXAttr is true (e.g., "user.sha256")
	XAttrChecksumKey string
	// If true and both files have source paths, prefer streaming file content from disk for checksum/byte compare
	StreamFromDiskIfAvailable bool
}

func defaultDiffOptions(caseSensitive bool) DiffOptions {
	return DiffOptions{
		CaseSensitive:   caseSensitive,
		ContentStrategy: CompareBytes,
		CompareMode:     true,
	}
}

// DiffWithOptions performs a diff with fine-grained options.
func DiffWithOptions(a, b *Folder, opts DiffOptions) op.Operation {
	return diffInternal(a, b, opts)
}

func Diff(a, b *Folder, caseSensitive bool) op.Operation {
	return diffInternal(a, b, defaultDiffOptions(caseSensitive))
}

// New: unified-config diff
func DiffWithConfig(a, b *Folder, cfg Config) op.Operation {
	// map Config to DiffOptions
	var strategy FileContentStrategy
	switch cfg.Strategy {
	case StructureOnly:
		strategy = SkipContent
	case Bytes:
		strategy = CompareBytes
	case ChecksumPrefer:
		strategy = PreferChecksumOrBytes
	case ChecksumEnsure:
		strategy = RequireChecksum // we will ensure below and not fallback to bytes
	case ChecksumRequire:
		strategy = RequireChecksum
	default:
		strategy = CompareBytes
	}
	opts := DiffOptions{
		CaseSensitive: cfg.CaseSensitive,
		ContentStrategy: strategy,
		ChecksumAlgorithm: cfg.Algorithm,
		CompareMode: cfg.CompareMode,
		CompareSize: cfg.CompareSize,
		CompareMTime: cfg.CompareMTime,
		ComputeChecksumIfMissing: cfg.Strategy == ChecksumPrefer || cfg.Strategy == ChecksumEnsure,
		WriteComputedChecksumToXAttr: false,
		StreamFromDiskIfAvailable: true,
	}
	return diffInternalWithExcludes(a, b, opts, cfg.ExcludeGlobs, cfg.ExcludeGlobs, "")
}

// Backwards-compatible wrapper without excludes
func diffInternal(a, b *Folder, opts DiffOptions) op.Operation {
	return diffInternalWithExcludes(a, b, opts, nil, nil, "")
}

func diffInternalWithExcludes(a, b *Folder, opts DiffOptions, aEx, bEx []string, prefix string) op.Operation {
	// if exclude globs differ, raise error by returning a Change op with a Reason
	if !sameStringSet(aEx, bEx) {
		return op.Operation{Operand: op.ChangeFolder, RelativePath: prefix, Value: op.DirValue{Reason: op.Reason{Type: op.Because, Before: aEx, After: bEx}}}
	}

	dirValue := op.DirValue{}

	a_index := 0
	b_index := 0
	a_keys := a.Entries()
	b_keys := b.Entries()

	if !opts.CaseSensitive {
		sortStringsToLower(a_keys)
		sortStringsToLower(b_keys)
	}

	for a_index < len(a_keys) && b_index < len(b_keys) {
		a_key := a_keys[a_index]
		b_key := b_keys[b_index]

		// exclude filtered entries
		if shouldExclude(normalizePath(prefix, a_key), aEx) {
			a_index++
			continue
		}
		if shouldExclude(normalizePath(prefix, b_key), bEx) {
			b_index++
			continue
		}

		a_comparable := a_key
		b_comparable := b_key

		if !opts.CaseSensitive {
			a_comparable = strings.ToLower(a_comparable)
			b_comparable = strings.ToLower(b_comparable)
		}

		if a_comparable == b_comparable {
			a_entry := a.Get(a_key)
			b_entry := b.Get(b_key)

			a_type := a_entry.Type()
			b_type := b_entry.Type()

			if a_type == FILE && b_type == FILE {
				changed, reason := filesDifferWithReason(a_entry.(*File), b_entry.(*File), opts)
				if changed {
					dirValue.AddOperations(a_entry.ChangeOperation(b_key, reason))
				}
			} else {
				equal, reason := a_entry.EqualWithReason(b_entry)
				if equal {
					// do nothing
				} else if a_type == FOLDER && b_type == FOLDER {
					a_entry := a_entry.(*Folder)
					b_entry := b_entry.(*Folder)
					operation := diffInternalWithExcludes(a_entry, b_entry, opts, aEx, bEx, normalizePath(prefix, b_key))
					operation.RelativePath = b_key
					if operation.Operand != op.Noop {
						dirValue.AddOperations(operation)
					}
				} else if a_type == FILE && b_type == FILE {
					// handled above
				} else {
					dirValue.AddOperations(
						a_entry.RemoveOperation(b_key, reason),
						b_entry.CreateOperation(b_key, reason),
					)
				}
			}

			a_index++
			b_index++
		} else if a_key < b_key {
			if shouldExclude(normalizePath(prefix, a_key), aEx) {
				a_index++
				continue
			}
			a_index++
			dirValue.AddOperations(a.RemoveChildOperation(a_key, op.Reason{}))
		} else if a_key > b_key {
			if shouldExclude(normalizePath(prefix, b_key), bEx) {
				b_index++
				continue
			}
			b_index++
			dirValue.AddOperations(b.CreateChildOperation(b_key, op.Reason{}))
		} else {
			panic("fsdt/diff.go(unreachable)")
		}
	}

	for a_index < len(a_keys) {
		relative_path := a_keys[a_index]
		a_index++
		if shouldExclude(normalizePath(prefix, relative_path), aEx) { continue }
		dirValue.AddOperations(a.RemoveChildOperation(relative_path, op.Reason{}))
	}
	for b_index < len(b_keys) {
		relative_path := b_keys[b_index]
		b_index++
		if shouldExclude(normalizePath(prefix, relative_path), bEx) { continue }
		dirValue.AddOperations(b.CreateChildOperation(relative_path, op.Reason{}))
	}

	if len(dirValue.Operations) == 0 {
		return op.Nothing
	}
	result := a.ChangeOperation(".", op.Reason{})
	result.Value = dirValue
	return result
}

func filesDifferWithReason(a, b *File, opts DiffOptions) (bool, op.Reason) {
	// First, check metadata if requested
	if changed, reason := fileMetadataDiff(a, b, opts); changed {
		return true, reason
	}

	switch opts.ContentStrategy {
	case SkipContent:
		return false, op.Reason{}
	case RequireChecksum:
		ad, an, aok := a.Checksum()
		bd, bn, bok := b.Checksum()
		if (!aok || !bok) && opts.ComputeChecksumIfMissing && opts.ChecksumAlgorithm != "" {
			ad, an, aok = ensureChecksum(a, ad, an, aok, opts)
			bd, bn, bok = ensureChecksum(b, bd, bn, bok, opts)
		}
		if !aok || !bok {
			// incompatible: missing required checksum
			return true, op.Reason{Type: op.Because, Before: "missing checksum", After: "missing checksum"}
		}
		if opts.ChecksumAlgorithm != "" && (an != opts.ChecksumAlgorithm || bn != opts.ChecksumAlgorithm) {
			return true, op.Reason{Type: op.Because, Before: an, After: bn}
		}
		if bytesEqual(ad, bd) {
			return false, op.Reason{}
		}
		return true, op.Reason{Type: op.ContentChanged, Before: ad, After: bd}
	case PreferChecksumOrBytes:
		ad, an, aok := a.Checksum()
		bd, bn, bok := b.Checksum()
		if (!aok || !bok) && opts.ComputeChecksumIfMissing && opts.ChecksumAlgorithm != "" {
			ad, an, aok = ensureChecksum(a, ad, an, aok, opts)
			bd, bn, bok = ensureChecksum(b, bd, bn, bok, opts)
		}
		if aok && bok && (opts.ChecksumAlgorithm == "" || (an == opts.ChecksumAlgorithm && bn == opts.ChecksumAlgorithm)) {
			if bytesEqual(ad, bd) {
				return false, op.Reason{}
			}
			return true, op.Reason{Type: op.ContentChanged, Before: ad, After: bd}
		}
		fallthrough
	case CompareBytes:
		// Compare bytes only if bytes mode
		if opts.ContentStrategy != CompareBytes {
			// do not fallback to bytes when strategy is not bytes-preferred
			return true, op.Reason{Type: op.Because, Before: "no checksum", After: "no checksum"}
		}
		if string(a.content) == string(b.content) {
			return false, op.Reason{}
		}
		return true, op.Reason{Type: op.ContentChanged, Before: a.content, After: b.content}
	default:
		if string(a.content) == string(b.content) {
			return false, op.Reason{}
		}
		return true, op.Reason{Type: op.ContentChanged, Before: a.content, After: b.content}
	}
}

func ensureChecksum(f *File, d []byte, n string, ok bool, opts DiffOptions) ([]byte, string, bool) {
	if ok {
		return d, n, ok
	}
	var content []byte
	if opts.StreamFromDiskIfAvailable {
		if path, has := f.SourcePath(); has {
			if data, err := readAllStreaming(path); err == nil {
				content = data
			}
		}
	}
	if content == nil {
		content = f.content
	}
	d = computeChecksum(opts.ChecksumAlgorithm, content)
	if d != nil {
		f.SetChecksum(opts.ChecksumAlgorithm, d)
		if opts.WriteComputedChecksumToXAttr {
			if path, has := f.SourcePath(); has && opts.XAttrChecksumKey != "" {
				_ = writeXAttrChecksum(path, opts.XAttrChecksumKey, d)
			}
		}
		return d, opts.ChecksumAlgorithm, true
	}
	return nil, n, false
}

func fileMetadataDiff(a, b *File, opts DiffOptions) (bool, op.Reason) {
	if opts.CompareMode && a.mode != b.mode {
		return true, op.Reason{Type: op.ModeChanged, Before: a.mode, After: b.mode}
	}
	if opts.CompareSize && a.size != b.size {
		return true, op.Reason{Type: op.SizeChanged, Before: a.size, After: b.size}
	}
	if opts.CompareMTime && !a.mtime.Equal(b.mtime) {
		return true, op.Reason{Type: op.MTimeChanged, Before: a.mtime, After: b.mtime}
	}
	return false, op.Reason{}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// DiffStreamWithConfig performs a diff and emits leaf operations as they are discovered.
// Each emitted operation has its RelativePath normalized to include the full relative path from root.
// Directory wrapper operations (ChangeDir) are not emitted; only leaf create/change/remove ops are streamed.
func DiffStreamWithConfig(a, b *Folder, cfg Config, emit func(op.Operation), visitDir func(path string)) {
	var strategy FileContentStrategy
	switch cfg.Strategy {
	case StructureOnly:
		strategy = SkipContent
	case Bytes:
		strategy = CompareBytes
	case ChecksumPrefer:
		strategy = PreferChecksumOrBytes
	case ChecksumEnsure:
		strategy = RequireChecksum // we will ensure below and not fallback to bytes
	case ChecksumRequire:
		strategy = RequireChecksum
	default:
		strategy = CompareBytes
	}
	opts := DiffOptions{
		CaseSensitive:            cfg.CaseSensitive,
		ContentStrategy:          strategy,
		ChecksumAlgorithm:        cfg.Algorithm,
		CompareMode:              cfg.CompareMode,
		CompareSize:              cfg.CompareSize,
		CompareMTime:             cfg.CompareMTime,
		ComputeChecksumIfMissing: cfg.Strategy == ChecksumPrefer || cfg.Strategy == ChecksumEnsure,
		WriteComputedChecksumToXAttr: false,
		StreamFromDiskIfAvailable:    true,
	}
	diffStreamInternalWithExcludes(a, b, opts, cfg.ExcludeGlobs, cfg.ExcludeGlobs, "", emit, visitDir)
}

func diffStreamInternalWithExcludes(a, b *Folder, opts DiffOptions, aEx, bEx []string, prefix string, emit func(op.Operation), visitDir func(path string)) {
	if visitDir != nil {
		visitDir(prefix)
	}

	a_index := 0
	b_index := 0
	a_keys := a.Entries()
	b_keys := b.Entries()

	if !opts.CaseSensitive {
		sortStringsToLower(a_keys)
		sortStringsToLower(b_keys)
	}

	for a_index < len(a_keys) && b_index < len(b_keys) {
		a_key := a_keys[a_index]
		b_key := b_keys[b_index]

		// exclude filtered entries
		if shouldExclude(normalizePath(prefix, a_key), aEx) {
			a_index++
			continue
		}
		if shouldExclude(normalizePath(prefix, b_key), bEx) {
			b_index++
			continue
		}

		a_comparable := a_key
		b_comparable := b_key

		if !opts.CaseSensitive {
			a_comparable = strings.ToLower(a_comparable)
			b_comparable = strings.ToLower(b_comparable)
		}

		if a_comparable == b_comparable {
			a_entry := a.Get(a_key)
			b_entry := b.Get(b_key)

			a_type := a_entry.Type()
			b_type := b_entry.Type()

			if a_type == FILE && b_type == FILE {
				changed, reason := filesDifferWithReason(a_entry.(*File), b_entry.(*File), opts)
				if changed {
					opr := a_entry.ChangeOperation(b_key, reason)
					opr.RelativePath = normalizePath(prefix, opr.RelativePath)
					emit(opr)
				}
			} else {
				equal, reason := a_entry.EqualWithReason(b_entry)
				if equal {
					// do nothing
				} else if a_type == FOLDER && b_type == FOLDER {
					a_entry := a_entry.(*Folder)
					b_entry := b_entry.(*Folder)
					diffStreamInternalWithExcludes(a_entry, b_entry, opts, aEx, bEx, normalizePath(prefix, b_key), emit, visitDir)
				} else if a_type == FILE && b_type == FILE {
					// handled above
				} else {
					emit(a_entry.RemoveOperation(normalizePath(prefix, b_key), reason))
					emit(b_entry.CreateOperation(normalizePath(prefix, b_key), reason))
				}
			}

			a_index++
			b_index++
		} else if a_key < b_key {
			if shouldExclude(normalizePath(prefix, a_key), aEx) {
				a_index++
				continue
			}
			a_index++
			emit(a.RemoveChildOperation(normalizePath(prefix, a_key), op.Reason{}))
		} else if a_key > b_key {
			if shouldExclude(normalizePath(prefix, b_key), bEx) {
				b_index++
				continue
			}
			b_index++
			emit(b.CreateChildOperation(normalizePath(prefix, b_key), op.Reason{}))
		} else {
			panic("fsdt/diff.go(unreachable)")
		}
	}

	for a_index < len(a_keys) {
		relative_path := a_keys[a_index]
		a_index++
		if shouldExclude(normalizePath(prefix, relative_path), aEx) { continue }
		emit(a.RemoveChildOperation(normalizePath(prefix, relative_path), op.Reason{}))
	}
	for b_index < len(b_keys) {
		relative_path := b_keys[b_index]
		b_index++
		if shouldExclude(normalizePath(prefix, relative_path), bEx) { continue }
		emit(b.CreateChildOperation(normalizePath(prefix, relative_path), op.Reason{}))
	}
}
