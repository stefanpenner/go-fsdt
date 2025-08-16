package fsdt

import (
	"encoding/hex"
	"os"
	"path/filepath"
)

type ChecksumStore interface {
	Load(path string) ([]byte, bool)
	Save(path string, sum []byte)
}

type XAttrStore struct{ Key string }

func (s XAttrStore) Load(path string) ([]byte, bool) {
	d, ok, _ := readXAttrChecksum(path, s.Key)
	return d, ok
}
func (s XAttrStore) Save(path string, sum []byte) { _ = writeXAttrChecksum(path, s.Key, sum) }

type SidecarStore struct{ BaseDir, Root string; Algorithm string }

func (s SidecarStore) cachePath(path string) (string, bool) {
	rel, err := filepath.Rel(s.Root, path)
	if err != nil {
		return "", false
	}
	return filepath.Join(s.BaseDir, rel+"."+s.Algorithm), true
}

func (s SidecarStore) Load(path string) ([]byte, bool) {
	p, ok := s.cachePath(path)
	if !ok { return nil, false }
	b, err := os.ReadFile(p)
	if err != nil { return nil, false }
	sum, err := hex.DecodeString(string(b))
	if err != nil { return nil, false }
	return sum, true
}

func (s SidecarStore) Save(path string, sum []byte) {
	p, ok := s.cachePath(path)
	if !ok { return }
	_ = os.MkdirAll(filepath.Dir(p), 0755)
	_ = os.WriteFile(p, []byte(hex.EncodeToString(sum)), 0644)
}

type MultiStore struct{ Stores []ChecksumStore }

func (m MultiStore) Load(path string) ([]byte, bool) {
	for _, s := range m.Stores {
		if d, ok := s.Load(path); ok { return d, true }
	}
	return nil, false
}
func (m MultiStore) Save(path string, sum []byte) {
	for _, s := range m.Stores { s.Save(path, sum) }
}