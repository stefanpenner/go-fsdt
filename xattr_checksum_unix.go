//go:build unix || linux || darwin

package fsdt

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"hash"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

func computeChecksum(algorithm string, data []byte) []byte {
	var h hash.Hash
	switch algorithm {
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	case "sha1":
		h = sha1.New()
	default:
		return nil
	}
	_, _ = io.Copy(h, bytesReader(data))
	return h.Sum(nil)
}

func bytesReader(b []byte) *bytesReaderImpl { return &bytesReaderImpl{b: b} }

type bytesReaderImpl struct{ b []byte }

func (r *bytesReaderImpl) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}

func readXAttrChecksum(path, key string) ([]byte, bool, error) {
	// Determine size
	sz, err := unix.Getxattr(path, key, nil)
	if err != nil {
		if errors.Is(err, unix.ENOATTR) || errors.Is(err, unix.ENODATA) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if sz <= 0 {
		return nil, false, nil
	}
	buf := make([]byte, sz)
	n, err := unix.Getxattr(path, key, buf)
	if err != nil {
		if errors.Is(err, unix.ENOATTR) || errors.Is(err, unix.ENODATA) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return buf[:n], true, nil
}

func writeXAttrChecksum(path, key string, value []byte) error {
	return unix.Setxattr(path, key, value, 0)
}