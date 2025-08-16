package fsdt

import (
	"path/filepath"
	"testing"

	op "github.com/stefanpenner/go-fsdt/operation"
	"github.com/stretchr/testify/require"
)

func Test_Diff_Scenarios(t *testing.T) {
	type scenario struct {
		name           string
		left           map[string]string
		right          map[string]string
		buildCfg       func(t *testing.T) Config
		expectNoChange bool
		assert         func(t *testing.T, d op.Operation)
	}

	scenarios := []scenario{
		{
			name: "Fast ignores content",
			left: map[string]string{"a.txt": "hello"},
			right: map[string]string{"a.txt": "world"},
			buildCfg: func(t *testing.T) Config { return DefaultFast() },
			expectNoChange: true,
		},
		{
			name: "Accurate detects content",
			left: map[string]string{"a.txt": "hello"},
			right: map[string]string{"a.txt": "world"},
			buildCfg: func(t *testing.T) Config { return DefaultAccurate() },
			expectNoChange: false,
		},
		{
			name: "ChecksumPrefer fallback to bytes: equal",
			left: map[string]string{"a.txt": "same content"},
			right: map[string]string{"a.txt": "same content"},
			buildCfg: func(t *testing.T) Config {
				cfg := DefaultAccurate()
				cfg.Strategy = ChecksumPrefer
				cfg.Algorithm = ""
				return cfg
			},
			expectNoChange: true,
		},
		{
			name: "ChecksumPrefer fallback to bytes: different (explain shows lengths)",
			left: map[string]string{"a.txt": "aaaaaaaaaa"},
			right: map[string]string{"a.txt": "bbbbbbbbbb"},
			buildCfg: func(t *testing.T) Config {
				cfg := DefaultAccurate()
				cfg.Strategy = ChecksumPrefer
				cfg.Algorithm = ""
				return cfg
			},
			expectNoChange: false,
			assert: func(t *testing.T, d op.Operation) {
				require.Contains(t, op.Explain(d), "content differs (len before 10, after 10)")
			},
		},
		{
			name: "ChecksumEnsure mode with sidecar store detects difference",
			left: map[string]string{"a.txt": "hello"},
			right: map[string]string{"a.txt": "world"},
			buildCfg: func(t *testing.T) Config {
				dir := t.TempDir()
				root := filepath.Join(dir, "root")
				side := filepath.Join(dir, "cache")
				store := MultiStore{Stores: []ChecksumStore{SidecarStore{BaseDir: side, Root: root, Algorithm: "sha256"}}}
				cfg := Checksums("sha256", store)
				cfg.Strategy = ChecksumEnsure
				return cfg
			},
			expectNoChange: false,
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			left := FS(sc.left)
			right := FS(sc.right)
			cfg := sc.buildCfg(t)
			d := DiffWithConfig(left, right, cfg)
			if sc.expectNoChange {
				require.Equal(t, op.Nothing, d)
			} else {
				require.NotEqual(t, op.Nothing, d)
			}
			if sc.assert != nil { sc.assert(t, d) }
		})
	}
}