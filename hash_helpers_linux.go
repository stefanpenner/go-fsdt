//go:build linux

package fsdt

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
)

func sha256New() hash.Hash { return sha256.New() }
func sha512New() hash.Hash { return sha512.New() }
func sha1New() hash.Hash   { return sha1.New() }