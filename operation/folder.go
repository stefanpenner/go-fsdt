package operation

type DirValue struct {
	Reason     Reason
	Operations []Operation
}

func NewRmdir(relativePath string, operations ...Operation) Operation {
	if len(operations) == 0 {
		return Operation{
			Operand:      Rmdir,
			RelativePath: relativePath,
			Value:        DirValue{},
		}
	} else {
		return Operation{
			Operand:      Rmdir,
			RelativePath: relativePath,
			Value: DirValue{
				Operations: operations,
			},
		}
	}
}

func NewChangeFolderOperation(relativePath string, operations ...Operation) Operation {
	return Operation{
		Operand:      ChangeFolder,
		RelativePath: relativePath,
		Value: DirValue{
			Operations: operations,
		},
	}
}

func NewMkdirOperation(relativePath string, operations ...Operation) Operation {
	if len(operations) == 0 {
		return Operation{
			Operand:      Mkdir,
			RelativePath: relativePath,
			Value:        DirValue{},
		}
	} else {
		return Operation{
			Operand:      Mkdir,
			RelativePath: relativePath,
			Value: DirValue{
				Operations: operations,
			},
		}
	}
}
