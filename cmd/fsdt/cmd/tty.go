package cmd

import (
	"os"
)

func isTerminal(f *os.File) bool {
	// Minimal, cross-platform-ish check; Bubble Tea will handle most rendering.
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func shouldEnableProgress(flag string) bool {
	switch flag {
	case "on":
		return true
	case "off":
		return false
	default:
		return isTerminal(os.Stderr)
	}
}