package operation

import "fmt"

// OperationType represents the type of file system operation
type OperationType string

// File system operation constants
const (
	// File operations
	CreateFile OperationType = "CreateFile"
	RemoveFile OperationType = "RemoveFile"
	ChangeFile OperationType = "ChangeFile"

	// Directory operations
	CreateDir OperationType = "CreateDir"
	RemoveDir OperationType = "RemoveDir"
	ChangeDir OperationType = "ChangeDir"

	// Link operations
	CreateLink OperationType = "CreateLink"
	RemoveLink OperationType = "RemoveLink"
	ChangeLink OperationType = "ChangeLink"

	// Special operations
	Noop OperationType = "Noop"
)

// Operation represents a file system operation
type Operation struct {
	RelativePath string
	Type         OperationType
	Value        OperationValue
	Reason       *Reason
}

// OperationValue represents the value associated with an operation
type OperationValue interface {
	// String returns a string representation of the value
	String() string
}

// Reason explains why an operation is needed
type Reason struct {
	Type    ReasonType
	Before  interface{}
	After   interface{}
	Message string
}

// ReasonType categorizes the reason for an operation
type ReasonType string

const (
	ReasonTypeChanged    ReasonType = "TypeChanged"
	ReasonModeChanged    ReasonType = "ModeChanged"
	ReasonContentChanged ReasonType = "ContentChanged"
	ReasonMissing        ReasonType = "Missing"
	ReasonNew            ReasonType = "New"
	ReasonPermissions    ReasonType = "Permissions"
	ReasonTimestamps     ReasonType = "Timestamps"
	ReasonAttributes     ReasonType = "Attributes"
)

// NewOperation creates a new operation with the given parameters
func NewOperation(relativePath string, opType OperationType, value OperationValue, reason *Reason) Operation {
	return Operation{
		RelativePath: relativePath,
		Type:         opType,
		Value:        value,
		Reason:       reason,
	}
}

// NewNoopOperation creates a no-operation
func NewNoopOperation() Operation {
	return Operation{
		Type: Noop,
	}
}

// IsNoop returns true if this is a no-operation
func (o Operation) IsNoop() bool {
	return o.Type == Noop
}

// String returns a string representation of the operation
func (o Operation) String() string {
	if o.IsNoop() {
		return "Noop"
	}

	result := fmt.Sprintf("%s: %s", o.Type, o.RelativePath)

	if o.Value != nil {
		result += fmt.Sprintf(" (%s)", o.Value.String())
	}

	if o.Reason != nil {
		result += fmt.Sprintf(" [%s]", o.Reason.String())
	}

	return result
}

// String returns a string representation of the reason
func (r *Reason) String() string {
	if r == nil {
		return ""
	}

	result := string(r.Type)

	if r.Message != "" {
		result += ": " + r.Message
	}

	if r.Before != nil || r.After != nil {
		result += " ("
		if r.Before != nil {
			result += fmt.Sprintf("before: %v", r.Before)
		}
		if r.Before != nil && r.After != nil {
			result += ", "
		}
		if r.After != nil {
			result += fmt.Sprintf("after: %v", r.After)
		}
		result += ")"
	}

	return result
}
