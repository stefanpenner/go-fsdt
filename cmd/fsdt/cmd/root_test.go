package cmd

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// helper to execute the root command with args and capture output
func executeCapture(args ...string) (string, string, error) {
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	rootCmd.SetArgs(args)
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&errBuf)
	err := rootCmd.Execute()
	return outBuf.String(), errBuf.String(), err
}

// captureStdout captures all writes to os.Stdout during fn execution.
func captureStdout(fn func() error) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	err := fn()
	_ = w.Close()
	os.Stdout = old
	out := <-done
	return out, err
}

func writeFile(t *testing.T, dir, name, content string, mtime time.Time) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { t.Fatal(err) }
	if err := os.WriteFile(path, []byte(content), fs.FileMode(0o644)); err != nil { t.Fatal(err) }
	if !mtime.IsZero() {
		if err := os.Chtimes(path, mtime, mtime); err != nil { t.Fatal(err) }
	}
	return path
}

func Test_CLI_NoMTime_Suppresses_MTimeOnly_Diffs(t *testing.T) {
	req := require.New(t)
	dir := t.TempDir()
	left := filepath.Join(dir, "left")
	right := filepath.Join(dir, "right")
	req.NoError(os.MkdirAll(left, 0o755))
	req.NoError(os.MkdirAll(right, 0o755))

	// same content, different mtime
	writeFile(t, left, "a.txt", "same", time.Unix(1000, 0))
	writeFile(t, right, "a.txt", "same", time.Unix(2000, 0))

	// default accurate mode should detect a change due to mtime
	out1, err := captureStdout(func() error {
		rootCmd.SetArgs([]string{"--format", "paths", left, right})
		return rootCmd.Execute()
	})
	req.NoError(err)
	req.NotEmpty(strings.TrimSpace(out1), "expected a changed path due to mtime difference")
	req.Contains(out1, "a.txt")

	// with --no-mtime, the output should be empty for paths format
	out2, err := captureStdout(func() error {
		rootCmd.SetArgs([]string{"--format", "paths", "--no-mtime", left, right})
		return rootCmd.Execute()
	})
	req.NoError(err)
	req.Equal("", strings.TrimSpace(out2))
}

// Sanity check that root command exists
func Test_RootCmd_Exists(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}
	if _, ok := interface{}(rootCmd).(*cobra.Command); !ok {
		t.Fatal("rootCmd is not a *cobra.Command")
	}
}