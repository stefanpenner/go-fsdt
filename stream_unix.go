//go:build unix || linux || darwin

package fsdt

import (
	"io"
	"os"
)

func readAllStreaming(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// streamEqualByPath compares two files by streaming from disk when source paths are available.
// Returns nil if streaming not possible; otherwise pointer to bool indicating equality.
func streamEqualByPath(a, b *File) *bool {
	ap, aok := a.SourcePath()
	bp, bok := b.SourcePath()
	if !aok || !bok {
		return nil
	}
	fa, err := os.Open(ap)
	if err != nil {
		return nil
	}
	defer fa.Close()
	fb, err := os.Open(bp)
	if err != nil {
		return nil
	}
	defer fb.Close()

	bufA := make([]byte, 64*1024)
	bufB := make([]byte, 64*1024)
	for {
		nA, eA := fa.Read(bufA)
		nB, eB := fb.Read(bufB)
		if nA != nB {
			res := false
			return &res
		}
		if nA > 0 && !bytesEqual(bufA[:nA], bufB[:nB]) {
			res := false
			return &res
		}
		if eA == io.EOF && eB == io.EOF {
			res := true
			return &res
		}
		if eA != nil && eA != io.EOF {
			return nil
		}
		if eB != nil && eB != io.EOF {
			return nil
		}
	}
}