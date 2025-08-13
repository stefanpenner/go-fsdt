package operation

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TreeFormatter formats operations as a tree structure
type TreeFormatter struct {
	// Configuration options
	ShowReasons bool
	ShowValues  bool
	IndentSize  int
}

// NewTreeFormatter creates a new tree formatter with default settings
func NewTreeFormatter() *TreeFormatter {
	return &TreeFormatter{
		ShowReasons: true,
		ShowValues:  true,
		IndentSize:  2,
	}
}

// Format formats an operation as a tree string
func (tf *TreeFormatter) Format(op Operation) string {
	return tf.formatOperation(op, 0, true)
}

// formatOperation recursively formats an operation and its children
func (tf *TreeFormatter) formatOperation(op Operation, level int, isLast bool) string {
	if op.IsNoop() {
		return "Noop"
	}

	// Build the prefix
	prefix := tf.buildPrefix(level, isLast)

	// Build the main line
	result := prefix + fmt.Sprintf("%s: %s", op.Type, op.RelativePath)

	// Add value information if requested
	if tf.ShowValues && op.Value != nil {
		result += fmt.Sprintf(" (%s)", op.Value.String())
	}

	// Add reason information if requested
	if tf.ShowReasons && op.Reason != nil {
		result += fmt.Sprintf(" [%s]", op.Reason.String())
	}

	// Add child operations for directory operations
	if dirValue, ok := op.Value.(DirectoryValue); ok && len(dirValue.Operations) > 0 {
		for i, childOp := range dirValue.Operations {
			isLastChild := i == len(dirValue.Operations)-1
			result += "\n" + tf.formatOperation(childOp, level+1, isLastChild)
		}
	}

	return result
}

// buildPrefix builds the visual prefix for tree formatting
func (tf *TreeFormatter) buildPrefix(level int, isLast bool) string {
	if level == 0 {
		return ""
	}

	result := ""
	for i := 0; i < level-1; i++ {
		result += "│" + strings.Repeat(" ", tf.IndentSize)
	}

	if isLast {
		result += "└── "
	} else {
		result += "├── "
	}

	return result
}

// FormatSimple formats operations in a simple, flat format
func (tf *TreeFormatter) FormatSimple(ops []Operation) string {
	var result []string

	for _, op := range ops {
		line := fmt.Sprintf("%s: %s", op.Type, op.RelativePath)

		if tf.ShowValues && op.Value != nil {
			line += fmt.Sprintf(" (%s)", op.Value.String())
		}

		if tf.ShowReasons && op.Reason != nil {
			line += fmt.Sprintf(" [%s]", op.Reason.String())
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// FormatJSON formats operations as JSON for machine consumption
func (tf *TreeFormatter) FormatJSON(ops []Operation) (string, error) {
	type jsonOp struct {
		Type         string      `json:"type"`
		RelativePath string      `json:"relative_path"`
		Value        interface{} `json:"value,omitempty"`
		Reason       *Reason     `json:"reason,omitempty"`
	}

	var jsonOps []jsonOp
	for _, op := range ops {
		jsonOp := jsonOp{
			Type:         string(op.Type),
			RelativePath: op.RelativePath,
			Value:        op.Value,
			Reason:       op.Reason,
		}
		jsonOps = append(jsonOps, jsonOp)
	}

	jsonBytes, err := json.MarshalIndent(jsonOps, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
