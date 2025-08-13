package operation

import "fmt"

// DirectoryValue represents the value for directory operations
type DirectoryValue struct {
	Operations  []Operation
	Mode        uint32
	Permissions uint32
}

// String returns a string representation of the directory value
func (d DirectoryValue) String() string {
	if len(d.Operations) == 0 {
		return "empty directory"
	}
	return fmt.Sprintf("directory with %d operations", len(d.Operations))
}

// AddOperations adds operations to the directory value
func (d *DirectoryValue) AddOperations(operations ...Operation) {
	d.Operations = append(d.Operations, operations...)
}

// NewCreateDirOperation creates a new directory creation operation
func NewCreateDirOperation(relativePath string, operations ...Operation) Operation {
	var value DirectoryValue
	if len(operations) > 0 {
		value = DirectoryValue{
			Operations: operations,
		}
	}

	reason := &Reason{
		Type:    ReasonNew,
		Message: "Directory does not exist in target",
	}

	return NewOperation(relativePath, CreateDir, value, reason)
}

// NewRemoveDirOperation creates a new directory removal operation
func NewRemoveDirOperation(relativePath string, operations ...Operation) Operation {
	var value DirectoryValue
	if len(operations) > 0 {
		value = DirectoryValue{
			Operations: operations,
		}
	}

	reason := &Reason{
		Type:    ReasonMissing,
		Message: "Directory exists in source but not in target",
	}

	return NewOperation(relativePath, RemoveDir, value, reason)
}

// NewChangeDirOperation creates a new directory change operation
func NewChangeDirOperation(relativePath string, operations ...Operation) Operation {
	var value DirectoryValue
	if len(operations) > 0 {
		value = DirectoryValue{
			Operations: operations,
		}
	}

	reason := &Reason{
		Type:    ReasonContentChanged,
		Message: "Directory contents have changed",
	}

	return NewOperation(relativePath, ChangeDir, value, reason)
}
