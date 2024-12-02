package fsdt

import (
	"os"
)

type Hardlink struct {
	target string
	mode   os.FileMode
}

func NewHardlink(target string) *Hardlink {
	return &Hardlink{
		target: target,
		mode:   0777,
	}
}

func (h *Hardlink) Type() FolderEntryType {
	return HARDLINK
}

func (h *Hardlink) Target() string {
	return h.target
}

func (h *Hardlink) Mode() os.FileMode {
	return h.mode
}

func (h *Hardlink) Clone() FolderEntry {
	return NewHardlink(h.target)
}

func (h *Hardlink) CreateOperation(relativePath string) Operation {
	return NewCreateHardlink(relativePath, h.target)
}

func (h *Hardlink) RemoveOperation(relativePath string) Operation {
	return NewUnlink(relativePath)
}

func (h *Hardlink) Equal(entry FolderEntry) bool {
	if entry.Type() != HARDLINK {
		return false
	}
	other, ok := entry.(*Hardlink)
	if !ok {
		return false
	}
	return h.target == other.target
}

func (h *Hardlink) EqualWithReason(entry FolderEntry) (bool, Reason) {
	if entry.Type() != HARDLINK {
		return false, Reason{Type: TypeChanged, Before: HARDLINK, After: entry.Type()}
	}
	other, ok := entry.(*Hardlink)
	if !ok {
		return false, Reason{Type: TypeChanged, Before: HARDLINK, After: entry.Type()}
	}
	if h.target != other.target {
		return false, Reason{Type: ContentChanged, Before: h.target, After: other.target}
	}
	return true, Reason{}
}

func (h *Hardlink) WriteTo(location string) error {
	return os.Link(h.target, location)
}

func (h *Hardlink) HasContent() bool {
	return true
}

func (h *Hardlink) Content() []byte {
	return []byte(h.target)
}

func (h *Hardlink) ContentString() string {
	return h.target
}

func (h *Hardlink) Strings(prefix string) []string {
	return []string{prefix + " => " + h.target}
}
