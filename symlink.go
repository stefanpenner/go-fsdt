package fsdt

import (
	"os"
)

type Symlink struct {
	target string
	mode   os.FileMode
}

func NewSymlink(target string) *Symlink {
	return &Symlink{
		target: target,
		mode:   0777,
	}
}

func (s *Symlink) Type() FolderEntryType {
	return SYMLINK
}

func (s *Symlink) Target() string {
	return s.target
}

func (s *Symlink) Mode() os.FileMode {
	return s.mode
}

func (s *Symlink) Clone() FolderEntry {
	return NewSymlink(s.target)
}

func (f *Symlink) RemoveOperation(relativePath string) Operation {
	return NewUnlink(relativePath)
}

func (f *Symlink) CreateOperation(relativePath string) Operation {
	// TODO: implement
	return NewLink(relativePath, f.target, SYMBOLIC)
}

func (s *Symlink) Equal(entry FolderEntry) bool {
	if entry.Type() != SYMLINK {
		return false
	}
	other, ok := entry.(*Symlink)
	if !ok {
		return false
	}
	return s.target == other.target
}

func (s *Symlink) EqualWithReason(entry FolderEntry) (bool, Reason) {
	if entry.Type() != SYMLINK {
		return false, Reason{Type: TypeChanged, Before: SYMLINK, After: entry.Type()}
	}
	other, ok := entry.(*Symlink)
	if !ok {
		return false, Reason{Type: TypeChanged, Before: SYMLINK, After: entry.Type()}
	}
	if s.target != other.target {
		return false, Reason{Type: ContentChanged, Before: s.target, After: other.target}
	}
	return true, Reason{}
}

func (s *Symlink) WriteTo(location string) error {
	return os.Symlink(s.target, location)
}

func (s *Symlink) HasContent() bool {
	return true
}

func (s *Symlink) Content() []byte {
	return []byte(s.target)
}

func (s *Symlink) ContentString() string {
	return s.target
}

func (s *Symlink) Strings(prefix string) []string {
	return []string{prefix + " -> " + s.target}
}
