package fsdt

import (
	op "github.com/stefanpenner/go-fsdt/operation"
)

const inlineContentMax = 1024 // bytes

// reasonForContentChange returns a Reason that inlines small contents, and summarizes large ones.
func reasonForContentChange(aContent []byte, bContent []byte, aSize int64, bSize int64) op.Reason {
	if aSize <= inlineContentMax && bSize <= inlineContentMax && len(aContent) > 0 && len(bContent) > 0 {
		return op.Reason{Type: op.ContentChanged, Before: append([]byte(nil), aContent...), After: append([]byte(nil), bContent...)}
	}
	return op.Reason{Type: op.ContentChanged, Before: op.ContentSummary{Size: aSize}, After: op.ContentSummary{Size: bSize}}
}