package operation

type FileChangedValue struct {
	Reason Reason
}

func NewFileOperation(relativePath string) Operation {
	return Operation{Operand: Create, RelativePath: relativePath}
}

func FileChangedOperation(relativePath string, reason Reason) Operation {
	return Operation{
		Operand:      ChangeFile,
		RelativePath: relativePath,
		Value: FileChangedValue{
			Reason: reason,
		},
	}
}
