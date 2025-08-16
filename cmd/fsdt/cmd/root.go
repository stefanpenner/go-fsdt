package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	fsdt "github.com/stefanpenner/go-fsdt"
	op "github.com/stefanpenner/go-fsdt/operation"
)

type options struct {
	mode string
	algo string
	xattrKey string
	sidecar string
	chkCache string
	root string
	precompute bool
	caseInsensitive bool
	format string
	excludes []string
	noMtime bool
}

var rootOpts options

var rootCmd = &cobra.Command{
	Use:   "fsdt [flags] <left> <right>",
	Short: "Fast, configurable filesystem diffing",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		left := filepath.Clean(args[0])
		right := filepath.Clean(args[1])
		if rootOpts.chkCache != "" {
			rootOpts.sidecar = rootOpts.chkCache
		}
		if rootOpts.root == "" { rootOpts.root = left }

		// Build store chain
		var stores []fsdt.ChecksumStore
		if rootOpts.xattrKey != "" {
			stores = append(stores, fsdt.XAttrStore{Key: rootOpts.xattrKey})
		}
		if rootOpts.sidecar != "" {
			stores = append(stores, fsdt.SidecarStore{BaseDir: rootOpts.sidecar, Root: rootOpts.root, Algorithm: rootOpts.algo})
		}
		var store fsdt.ChecksumStore
		if len(stores) == 1 { store = stores[0] }
		if len(stores) > 1 { store = fsdt.MultiStore{Stores: stores} }

		// Load trees
		a := fsdt.NewFolder()
		b := fsdt.NewFolder()
		load := fsdt.LoadOptions{}
		if rootOpts.xattrKey != "" {
			load.XAttrChecksumKey = rootOpts.xattrKey
			load.ChecksumAlgorithm = rootOpts.algo
			load.ComputeChecksumIfMissing = false
			load.WriteComputedChecksumToXAttr = false
		}
		if err := a.ReadFromWithOptions(left, load); err != nil { return err }
		if err := b.ReadFromWithOptions(right, load); err != nil { return err }

		// Config
		cfg := fsdt.Config{CaseSensitive: !rootOpts.caseInsensitive}
		switch rootOpts.mode {
		case "fast": cfg = fsdt.DefaultFast()
		case "accurate": cfg = fsdt.DefaultAccurate()
		case "checksum": cfg = fsdt.Checksums(rootOpts.algo, store)
		case "checksum-ensure": cfg = fsdt.Checksums(rootOpts.algo, store); cfg.Strategy = fsdt.ChecksumEnsure
		case "checksum-require": cfg = fsdt.ChecksumsStrict(rootOpts.algo, store)
		default:
			return fmt.Errorf("unknown mode: %s", rootOpts.mode)
		}
		cfg.CaseSensitive = !rootOpts.caseInsensitive
		cfg.ExcludeGlobs = append([]string(nil), rootOpts.excludes...)
		// Apply mtime exclusion if requested (only disables, never enables)
		if rootOpts.noMtime {
			cfg.CompareMTime = false
		}

		// Precompute
		if rootOpts.precompute && store != nil && (cfg.Strategy == fsdt.ChecksumPrefer || cfg.Strategy == fsdt.ChecksumEnsure) {
			precomputeTreeChecksums(a, rootOpts.algo, store, left)
			precomputeTreeChecksums(b, rootOpts.algo, store, right)
		}

		d := fsdt.DiffWithConfig(a, b, cfg)
		if dv, ok := d.Value.(op.DirValue); ok && dv.Reason.Type == op.Because {
			return fmt.Errorf("incompatible or missing prerequisites: %v -> %v", dv.Reason.Before, dv.Reason.After)
		}

		switch rootOpts.format {
		case "pretty", "tree":
			fmt.Println(op.Print(d))
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(d)
		case "paths":
			for _, p := range collectPaths(d) { fmt.Println(p) }
		default:
			return fmt.Errorf("unknown format: %s", rootOpts.format)
		}
		return nil
	},
}

func init() {
	rootCmd.Flags().StringVar(&rootOpts.mode, "mode", "accurate", "diff mode: fast|accurate|checksum|checksum-ensure|checksum-require")
	rootCmd.Flags().StringVar(&rootOpts.algo, "algo", "sha256", "checksum algorithm for checksum modes (e.g., sha256)")
	rootCmd.Flags().StringVar(&rootOpts.xattrKey, "xattr", "", "xattr key (e.g., Linux: user.sha256; macOS: com.yourorg.sha256 or sha256)")
	rootCmd.Flags().StringVar(&rootOpts.sidecar, "sidecar", "", "external checksum cache dir (mirrors relative paths under --root with extension .<algo>)")
	rootCmd.Flags().StringVar(&rootOpts.chkCache, "checksum-cache-dir", "", "alias of --sidecar")
	rootCmd.Flags().StringVar(&rootOpts.root, "root", "", "project root for sidecar relative paths (defaults to left)")
	rootCmd.Flags().BoolVar(&rootOpts.precompute, "precompute", false, "precompute and persist missing checksums before diff (when using a store)")
	rootCmd.Flags().BoolVar(&rootOpts.caseInsensitive, "ci", false, "case-insensitive diff")
	rootCmd.Flags().StringVar(&rootOpts.format, "format", "pretty", "output format: pretty|tree|json|paths")
	rootCmd.Flags().StringArrayVar(&rootOpts.excludes, "exclude", nil, "exclude glob (repeatable), supports doublestar patterns")
	rootCmd.Flags().BoolVar(&rootOpts.noMtime, "no-mtime", false, "exclude mtime from comparison")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func precomputeTreeChecksums(folder *fsdt.Folder, algo string, store fsdt.ChecksumStore, root string) {
	opts := fsdt.ChecksumOptions{Algorithm: algo}
	switch s := store.(type) {
	case fsdt.XAttrStore:
		opts.XAttrKey = s.Key
		opts.WriteToXAttr = true
	case fsdt.SidecarStore:
		opts.SidecarDir = s.BaseDir
		opts.RootPath = s.Root
		opts.WriteToXAttr = false
	case fsdt.MultiStore:
		for _, inner := range s.Stores { precomputeTreeChecksums(folder, algo, inner, root) }
		return
	}
	opts.ComputeIfMissing = true
	opts.StreamFromDiskIfAvailable = true
	recursiveEnsure(folder, opts)
}

func recursiveEnsure(entry fsdt.FolderEntry, opts fsdt.ChecksumOptions) {
	switch v := entry.(type) {
	case *fsdt.File:
		_, _, _ = v.EnsureChecksum(opts)
	case *fsdt.Folder:
		for _, name := range v.Entries() {
			recursiveEnsure(v.Get(name), opts)
		}
		_, _, _ = v.EnsureChecksum(opts)
	}
}

func collectPaths(d op.Operation) []string {
	var out []string
	var walk func(op.Operation)
	walk = func(node op.Operation) {
		if node.Operand != op.Noop && node.RelativePath != "." {
			out = append(out, node.RelativePath)
		}
		if dv, ok := node.Value.(op.DirValue); ok {
			for _, c := range dv.Operations { walk(c) }
		}
	}
	walk(d)
	return out
}