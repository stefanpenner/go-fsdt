package operation

import "fmt"

type (
	Operand string
)

// TODO: Create -> CreateFile
const (
	Create       Operand = "CreateFile"
	ChangeFile   Operand = "ChangeFile"
	ChangeFolder Operand = "ChangeDir"
	Rmdir        Operand = "Rmdir"
	Mkdir        Operand = "Mkdir"
	Noop         Operand = "noop"
)

const (
	UI_T = "├── "
	UI_V = "│   "
	UI_J = "└── "
)

type Reason struct {
	Before interface{}
	After  interface{}
	Type   ReasonType
}

// TODDO: expand reason from enum, to struct (before / after)
type ReasonType string

const (
	TypeChanged    ReasonType = "Type Changed"
	ModeChanged    ReasonType = "Mode Changed"
	ContentChanged ReasonType = "Content Changed"
	Missing        ReasonType = "Missing"
	Because        ReasonType = "because"
	SizeChanged    ReasonType = "Size Changed"
	MTimeChanged   ReasonType = "MTime Changed"
)

type Operation struct {
	RelativePath string
	Value        Value
	Operand      Operand
}

var Nothing = Operation{
	Operand: Noop,
}

func prefix(level int, isLast bool) string {
	result := ""
	for i := 0; i < level; i++ {
		result += UI_V
	}

	if isLast {
		result += UI_J
	} else {
		result += UI_T
	}

	return result
}

func Print(op Operation) string {
	var isLast bool
	if value, ok := op.Value.(DirValue); ok {
		isLast = len(value.Operations) == 0
	} else {
		isLast = true
	}

	return print(op, 0, isLast)
}

func print(op Operation, level int, isLast bool) string {
	result := prefix(level, isLast)

	result += fmt.Sprintf("%s: ", op.Operand) + op.RelativePath

	if value, ok := op.Value.(DirValue); ok {
		length := len(value.Operations)
		for index, op := range value.Operations {
			result += "\n" + print(op, level+1, index >= length-1)
		}
	}

	return result
}
