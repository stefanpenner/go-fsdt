package fsdt

import (
	"regexp"
	"testing"

	op "github.com/stefanpenner/go-fsdt/operation"
	"github.com/stretchr/testify/require"
)

func Test_PreferChecksumOrBytes_FallbackToBytes_Equal(t *testing.T) {
	req := require.New(t)
	left := NewFolder()
	right := NewFolder()

	left.FileString("a.txt", "same content")
	right.FileString("a.txt", "same content")

	opts := DiffOptions{
		CaseSensitive: true,
		ContentStrategy: PreferChecksumOrBytes,
		ChecksumAlgorithm: "",              // no algorithm configured
		ComputeChecksumIfMissing: false,     // do not compute checksum -> must fallback to bytes
	}

	d := DiffWithOptions(left, right, opts)
	req.Equal(op.Nothing, d, "identical content must not produce a change when falling back to bytes")
}

func Test_PreferChecksumOrBytes_FallbackToBytes_Different(t *testing.T) {
	req := require.New(t)
	left := NewFolder()
	right := NewFolder()

	left.FileString("a.txt", "aaaaaaaaaa")
	right.FileString("a.txt", "bbbbbbbbbb") // same length, different content

	opts := DiffOptions{
		CaseSensitive: true,
		ContentStrategy: PreferChecksumOrBytes,
		ChecksumAlgorithm: "",
		ComputeChecksumIfMissing: false,
	}

	d := DiffWithOptions(left, right, opts)
	req.NotEqual(op.Nothing, d, "different content should be detected when falling back to bytes")

	// Human-readable explanation should indicate content differs and show equal lengths
	explain := op.Explain(d)
	req.Regexp(regexp.MustCompile(`content differs \(len before 10, after 10\)`), explain)
}

func Test_CompareBytes_Identical_NoChange(t *testing.T) {
	req := require.New(t)
	left := NewFolder()
	right := NewFolder()

	payload := "some data including \x00 bytes and unicode ✓"
	left.FileString("a.txt", payload)
	right.FileString("a.txt", payload)

	opts := DiffOptions{CaseSensitive: true, ContentStrategy: CompareBytes}
	d := DiffWithOptions(left, right, opts)
	req.Equal(op.Nothing, d)
}