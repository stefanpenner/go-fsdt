package operation

import (
	"fmt"
	"os"
	"time"
)

// Explain renders a human-readable explanation of why changes occurred.
// It preserves a tree-like structure similar to Print, but includes reasons.
func Explain(op Operation) string {
	var isLast bool
	if value, ok := op.Value.(DirValue); ok {
		isLast = len(value.Operations) == 0
	} else {
		isLast = true
	}
	return explain(op, 0, isLast)
}

func explain(op Operation, level int, isLast bool) string {
	result := prefix(level, isLast)
	result += fmt.Sprintf("%s: %s", op.Operand, op.RelativePath)

	// Append reason (if any)
	switch v := op.Value.(type) {
	case FileChangedValue:
		if v.Reason.Type != "" {
			result += " — " + formatReason(v.Reason)
		}
	case DirValue:
		if v.Reason.Type != "" {
			result += " — " + formatReason(v.Reason)
		}
		length := len(v.Operations)
		for idx, child := range v.Operations {
			result += "\n" + explain(child, level+1, idx >= length-1)
		}
	}

	return result
}

func formatReason(r Reason) string {
	switch r.Type {
	case ContentChanged:
		// Avoid dumping content; give concise summary
		bl, al := lengthOf(r.Before), lengthOf(r.After)
		if bl >= 0 && al >= 0 {
			return fmt.Sprintf("content differs (len before %d, after %d)", bl, al)
		}
		return "content differs"
	case ModeChanged:
		return fmt.Sprintf("mode changed (%s → %s)", formatFileMode(r.Before), formatFileMode(r.After))
	case SizeChanged:
		return fmt.Sprintf("size changed (%s → %s)", formatInt64(r.Before), formatInt64(r.After))
	case MTimeChanged:
		return fmt.Sprintf("mtime changed (%s → %s)", formatTime(r.Before), formatTime(r.After))
	case TypeChanged:
		return fmt.Sprintf("type changed (%v → %v)", r.Before, r.After)
	case Missing:
		return fmt.Sprintf("missing (%v)", r.Before)
	case Because:
		return fmt.Sprintf("because: %v → %v", r.Before, r.After)
	default:
		if r.Type != "" {
			return string(r.Type)
		}
		return ""
	}
}

func lengthOf(v interface{}) int {
	switch t := v.(type) {
	case []byte:
		return len(t)
	case string:
		return len(t)
	case ContentSummary:
		if t.Size < 0 { return -1 }
		if t.Size > int64(int(^uint(0)>>1)) { return int(^uint(0)>>1) }
		return int(t.Size)
	default:
		return -1
	}
}

func formatFileMode(v interface{}) string {
	if m, ok := v.(os.FileMode); ok {
		return fmt.Sprintf("%#o", m)
	}
	return fmt.Sprintf("%v", v)
}

func formatInt64(v interface{}) string {
	switch t := v.(type) {
	case int64:
		return fmt.Sprintf("%d", t)
	case int:
		return fmt.Sprintf("%d", t)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatTime(v interface{}) string {
	if tm, ok := v.(time.Time); ok {
		return tm.UTC().Format(time.RFC3339)
	}
	return fmt.Sprintf("%v", v)
}