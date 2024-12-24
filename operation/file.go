package operation

import "fmt"

type FileChangedValue struct {
	Reason Reason
}

func (f FileChangedValue) Print(indent string) string {
	// return fmt.Sprintf("%s %s", indent, f.Reason)
	return fmt.Sprintf("%s - %s", indent, "Because")
}

func NewFileOperation(relativePath string) Operation {
	return Operation{Operand: Create, RelativePath: relativePath}
}
