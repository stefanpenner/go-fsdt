package fsdt

import (
	"fmt"
	"os"

	op "github.com/stefanpenner/go-fsdt/operation"
)

type Link struct {
	target    string
	link_type FolderEntryType // only SYMLINK or HARDLINK
	mode      os.FileMode
}

func NewLink(target string, link_type FolderEntryType) *Link {
	if link_type == SYMLINK {
		return &Link{
			target:    target,
			mode:      0777,
			link_type: SYMLINK,
		}
	} else {
		panic("go-fsdt/NewLink only symlinks supported")
	}
}

func (l *Link) Type() FolderEntryType {
	return l.link_type
}

func (l *Link) Target() string {
	return l.target
}

func (l *Link) Mode() os.FileMode {
	return l.mode
}

func (l *Link) Clone() FolderEntry {
	return NewLink(l.target, l.link_type)
}

func (l *Link) RemoveOperation(relativePath string, reason op.Reason) op.Operation {
	// TODO: reason
	return op.NewUnlink(relativePath)
}

func (l *Link) OperationLinkType() op.LinkType {
	switch l.Type() {
	case SYMLINK:
		return op.SYMBOLIC_LINK
	case HARDLINK:
		return op.HARD_LINK
	default:
		panic("Invalid link type")
	}
}

func (l *Link) CreateOperation(relativePath string, reason op.Reason) op.Operation {
	// TODO: reason
	return op.NewCreateLink(relativePath, l.Target())
}

func (l *Link) ChangeOperation(relativePath string, reason op.Reason, operations ...op.Operation) op.Operation {
	panic("no implemented")
	return op.Operation{}
}

func (l *Link) Equal(entry FolderEntry) bool {
	other, ok := entry.(*Link)
	if !ok {
		return false
	}
	return l.target == other.target && l.link_type == other.link_type
}

func (l *Link) EqualWithReason(entry FolderEntry) (bool, op.Reason) {
	other, ok := entry.(*Link)
	if !ok {
		return false, op.Reason{Type: op.TypeChanged, Before: SYMLINK, After: entry.Type()}
	}
	if l.target != other.target {
		return false, op.Reason{Type: op.ContentChanged, Before: l.target, After: other.target}
	}
	if l.link_type != other.link_type {
		return false, op.Reason{Type: op.ContentChanged, Before: l.link_type, After: other.link_type}
	}
	return true, op.Reason{}
}

func (l *Link) WriteTo(link string) error {
	if l.Type() == SYMLINK {
		return os.Symlink(l.target, link)
	} else if l.Type() == HARDLINK {
		panic("go-fsdt/hardlink not supported")
	} else {
		return fmt.Errorf("unexpected link type: %s", l.Type())
	}
}

func (l *Link) HasContent() bool {
	return true
}

func (l *Link) Content() []byte {
	return []byte(l.target)
}

func (l *Link) ContentString() string {
	return l.target
}

func (l *Link) Strings(prefix string) []string {
	return []string{prefix + " -> " + l.target}
}
