package operation

import "fmt"

// LinkType represents the type of link
type LinkType string

const (
	HardLink     LinkType = "hardlink"
	SymbolicLink LinkType = "symbolic"
)

// LinkValue represents the value for link operations
type LinkValue struct {
	LinkType    LinkType
	Target      string
	Mode        uint32
	Permissions uint32
}

// String returns a string representation of the link value
func (l LinkValue) String() string {
	return fmt.Sprintf("%s link -> %s (mode: %o)", l.LinkType, l.Target, l.Mode)
}

// NewCreateLinkOperation creates a new link creation operation
func NewCreateLinkOperation(relativePath string, target string, linkType LinkType, mode uint32) Operation {
	value := LinkValue{
		LinkType: linkType,
		Target:   target,
		Mode:     mode,
	}

	reason := &Reason{
		Type:    ReasonNew,
		Message: "Link does not exist in target",
	}

	return NewOperation(relativePath, CreateLink, value, reason)
}

// NewRemoveLinkOperation creates a new link removal operation
func NewRemoveLinkOperation(relativePath string) Operation {
	reason := &Reason{
		Type:    ReasonMissing,
		Message: "Link exists in source but not in target",
	}

	return NewOperation(relativePath, RemoveLink, nil, reason)
}

// NewChangeLinkOperation creates a new link change operation
func NewChangeLinkOperation(relativePath string, before, after LinkValue, changeType ReasonType) Operation {
	value := LinkValue{
		LinkType: after.LinkType,
		Target:   after.Target,
		Mode:     after.Mode,
	}

	reason := &Reason{
		Type:    changeType,
		Before:  before,
		After:   after,
		Message: fmt.Sprintf("Link %s changed", changeType),
	}

	return NewOperation(relativePath, ChangeLink, value, reason)
}
