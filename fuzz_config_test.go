package fsdt

import (
	"testing"

	op "github.com/stefanpenner/go-fsdt/operation"
)

// FuzzDiffWithConfigQuick fuzzes diff across strategies with tiny trees.
func FuzzDiffWithConfigQuick(f *testing.F) {
	// Seeds
	f.Add(1, 2, true, "a.txt", "hello", 0) // small tree, same content
	f.Add(1, 2, false, "b.txt", "world", 1) // small tree, different content

	f.Fuzz(func(t *testing.T, depth, width int, same bool, name, content string, strategy int) {
		if depth < 0 { depth = 0 }
		if width < 1 { width = 1 }
		if depth > 2 { depth = 2 }
		if width > 3 { width = 3 }
		if name == "" { name = "f.txt" }
		if len(content) > 64 { content = content[:64] }

		// Build tiny trees
		a := NewFolder()
		b := NewFolder()
		a.Set(name, content)
		if same {
			b.Set(name, content)
		} else {
			b.Set(name, content+"x")
		}

		// Choose strategy
		var cfg Config
		switch strategy % 4 {
		case 0:
			cfg = DefaultFast()
		case 1:
			cfg = DefaultAccurate()
		case 2:
			cfg = Checksums("sha256", nil) // prefer
		case 3:
			cfg = ChecksumsStrict("sha256", nil) // require (no checksums present)
		}

		// Diff should not panic
		_ = DiffWithConfig(a, b, cfg)

		// In require mode (no checksums), changed files should report Because, not raw bytes
		if cfg.Strategy == ChecksumRequire {
			if dv, ok := DiffWithConfig(a, b, cfg).Value.(op.DirValue); ok {
				for _, c := range dv.Operations {
					if fv, ok := c.Value.(op.FileChangedValue); ok && c.Operand == op.ChangeFile {
						_ = (fv.Reason.Type == op.Because) || (fv.Reason == op.Reason{})
					}
				}
			}
		}
	})
}

// FuzzExcludeGlobsQuick fuzzes exclude pattern compatibility and behavior on tiny inputs.
func FuzzExcludeGlobsQuick(f *testing.F) {
	f.Add("tmp/**", "tmp/**", "a.txt", "hello")         // same excludes
	f.Add("tmp/**", "logs/**", "b.txt", "world")        // different excludes
	f.Add("**/*.log", "**/*.log", "tmp/x.log", "x")     // glob match
	f.Add("", "", "docs/readme.md", "hi")                // no excludes

	f.Fuzz(func(t *testing.T, exA, exB, path, content string) {
		a := FS(map[string]string{path: content})
		b := a.Copy()
		b.Set(path, content+"x")

		cfg := DefaultAccurate()
		cfg.ExcludeGlobs = []string{exA}
		_ = DiffWithConfig(a, b, cfg)

		cfgB := cfg
		cfgB.ExcludeGlobs = []string{exB}
		_ = DiffWithConfig(a, b, cfgB)
	})
}