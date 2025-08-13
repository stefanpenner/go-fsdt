package operation

type DirValue struct {
	Reason     Reason
	Operations []Operation
	// TODO: mode, permisssions
}

func (d *DirValue) AddOperations(operations ...Operation) {
	d.Operations = append(d.Operations, operations...)
}

func NewRmdir(relativePath string, operations ...Operation) Operation {
	var value DirValue
	if len(operations) != 0 {
		value = DirValue{
			Operations: make([]Operation, 0, len(operations)),
		}
		value.AddOperations(operations...)
	}
	return Operation{
		Operand:      Rmdir,
		RelativePath: relativePath,
		Value:        value,
	}
}

func NewChangeFolderOperation(relativePath string, operations ...Operation) Operation {
	return Operation{
		Operand:      ChangeFolder,
		RelativePath: relativePath,
		Value: DirValue{
			Operations: make([]Operation, 0, len(operations)),
		},
	}
}

func NewMkdirOperation(relativePath string, operations ...Operation) Operation {
	var value DirValue

	if len(operations) != 0 {
		value = DirValue{
			Operations: make([]Operation, 0, len(operations)),
		}
		value.AddOperations(operations...)
	}

	return Operation{
		Operand:      Mkdir,
		RelativePath: relativePath,
		Value:        value,
	}
}
