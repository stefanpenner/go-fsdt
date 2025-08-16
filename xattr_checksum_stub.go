//go:build !unix && !linux && !darwin

package fsdt

func computeChecksum(algorithm string, data []byte) []byte { return nil }

func readXAttrChecksum(path, key string) ([]byte, bool, error) { return nil, false, nil }

func writeXAttrChecksum(path, key string, value []byte) error { return nil }