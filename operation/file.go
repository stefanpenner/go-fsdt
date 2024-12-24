package operation

import "fmt"

type FileChangedValue struct {
	Reason Reason
}

func (f FileChangedValue) Print(indent string, prefix string) string {
	return fmt.Sprintf("%s - %s", indent, "Because")
}

func NewFileOperation(relativePath string) Operation {
	return Operation{Operand: Create, RelativePath: relativePath}
}

func NewChangeFileOperation(relativePath string) Operation {
	// TODO: reason
	return Operation{Operand: ChangeFile, RelativePath: relativePath}
}
