package operation

import "fmt"

type LinkType string

const (
	HARD_LINK     LinkType = "hardlink"
	SYMBOLIC_LINK LinkType = "symbolic"
)

type LinkValue struct {
	LinkType LinkType
	Target   string
}

const (
	Unlink     Operand = "Unlink"
	CreateLink Operand = "CreateLink"
)

func (l LinkValue) Print(indent string, prefix string) string {
	return fmt.Sprintf("%s%s -> %s", indent, prefix, l.Target)
}

func NewUnlink(relativePath string) Operation {
	return Operation{
		Operand:      Unlink,
		RelativePath: relativePath,
	}
}

func NewCreateLink(relativePath string, target string) Operation {
	return Operation{
		Operand:      CreateLink,
		RelativePath: relativePath,
		Value: LinkValue{
			LinkType: SYMBOLIC_LINK,
			Target:   target,
		},
	}
}
