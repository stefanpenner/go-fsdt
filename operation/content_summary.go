package operation

// ContentSummary carries lightweight info about content without embedding raw bytes.
// Size is the content length in bytes. DigestPrefix may optionally include a short
// hex-encoded prefix of a digest, and Algorithm names that digest.
type ContentSummary struct {
	Size          int64
	DigestPrefix  string
	Algorithm     string
}