package fsdt

import (
	"bytes"
	"hash"
	"io"
	"os"
)

func newHash(algorithm string) hash.Hash {
	switch algorithm {
	case "sha256":
		return sha256New()
	case "sha512":
		return sha512New()
	case "sha1":
		return sha1New()
	default:
		return nil
	}
}

func ioWriteString(h hash.Hash, s string) {
	_, _ = io.Copy(h, bytes.NewBufferString(s))
}

// computeChecksumFromPathOrBytes prefers streaming from path when available.
func computeChecksumFromPathOrBytes(algorithm, path string, data []byte) []byte {
	h := newHash(algorithm)
	if h == nil {
		return nil
	}
	if path != "" {
		f, err := os.Open(path)
		if err == nil {
			defer f.Close()
			_, _ = io.Copy(h, f)
			return h.Sum(nil)
		}
	}
	_, _ = io.Copy(h, bytes.NewReader(data))
	return h.Sum(nil)
}