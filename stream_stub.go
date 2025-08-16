//go:build !unix && !linux && !darwin

package fsdt

func readAllStreaming(path string) ([]byte, error) { return nil, nil }

func streamEqualByPath(a, b *File) *bool { return nil }