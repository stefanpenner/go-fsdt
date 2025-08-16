//go:build !linux && !darwin

package fsdt

import "hash"

func sha256New() hash.Hash { return nil }
func sha512New() hash.Hash { return nil }
func sha1New() hash.Hash   { return nil }