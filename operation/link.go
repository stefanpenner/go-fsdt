package operation

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

func NewUnlink(relativePath string) Operation {
	return Operation{
		Operand:      Unlink,
		RelativePath: relativePath,
	}
}

func NewCreateLink(relativePath string, target string, linkType LinkType) Operation {
	if linkType == SYMBOLIC_LINK {
		return Operation{
			Operand:      CreateLink,
			RelativePath: relativePath,
			Value: LinkValue{
				LinkType: linkType,
				Target:   target,
			},
		}
	} else {
		panic("cannot create NewCreateLink that isn't a symlink") // TODO: unify and provide a good error
	}
}
