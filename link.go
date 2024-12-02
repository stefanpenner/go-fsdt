package fsdt

import (
	"os"

	op "github.com/stefanpenner/go-fsdt/operation"
)

type Link struct {
	target    string
	link_type FolderEntryType // only SYMLINK or HARDLINK
	mode      os.FileMode
}

// TODO: explore generics for constraining, first glance made it look a little tedious
func NewLink(target string, link_type FolderEntryType) *Link {
	if link_type == HARDLINK || link_type == SYMLINK {
		return &Link{
			target:    target,
			mode:      0777,
			link_type: link_type,
		}
	} else {
		panic("New Link type must be either hardlink or symlink")
	}
}

func (s *Link) Type() FolderEntryType {
	return s.link_type
}

func (s *Link) Target() string {
	return s.target
}

func (s *Link) Mode() os.FileMode {
	return s.mode
}

func (s *Link) Clone() FolderEntry {
	return NewLink(s.target, s.link_type)
}

func (f *Link) RemoveOperation(relativePath string) op.Operation {
	return op.NewUnlink(relativePath)
}

func (f *Link) OperationLinkType() op.LinkType {
	switch f.Type() {
	case SYMLINK:
		return op.SYMBOLIC_LINK
	case HARDLINK:
		return op.HARD_LINK
	default:
		panic("Invalid link type")
	}
}

func (f *Link) CreateOperation(relativePath string) op.Operation {
	return op.NewCreateLink(relativePath, f.Target(), f.OperationLinkType())
}

func (s *Link) Equal(entry FolderEntry) bool {
	other, ok := entry.(*Link)
	if !ok {
		return false
	}
	return s.target == other.target && s.link_type == other.link_type
}

func (s *Link) EqualWithReason(entry FolderEntry) (bool, op.Reason) {
	other, ok := entry.(*Link)
	if !ok {
		return false, op.Reason{Type: op.TypeChanged, Before: SYMLINK, After: entry.Type()}
	}
	if s.target != other.target {
		return false, op.Reason{Type: op.ContentChanged, Before: s.target, After: other.target}
	}
	if s.link_type != other.link_type {
		return false, op.Reason{Type: op.ContentChanged, Before: s.link_type, After: other.link_type}
	}
	return true, op.Reason{}
}

func (s *Link) WriteTo(location string) error {
	return os.Link(s.target, location)
}

func (s *Link) HasContent() bool {
	return true
}

func (s *Link) Content() []byte {
	return []byte(s.target)
}

func (s *Link) ContentString() string {
	return s.target
}

func (s *Link) Strings(prefix string) []string {
	return []string{prefix + " -> " + s.target}
}
