package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	cerrs "github.com/cockroachdb/errors"
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
	progress string
}

var rootOpts options

var rootCmd = &cobra.Command{
	Use:   "fsdt [flags] <left> <right>",
	Short: "Fast, configurable filesystem diffing",
	Args:  cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		left := filepath.Clean(args[0])
		right := filepath.Clean(args[1])

		// Normalize aliases
		if rootOpts.chkCache != "" {
			rootOpts.sidecar = rootOpts.chkCache
		}
		if rootOpts.root == "" {
			rootOpts.root = left
		}

		// Validate flags
		switch rootOpts.mode {
		case "fast", "accurate", "checksum", "checksum-ensure", "checksum-require":
		default:
			return cerrs.Newf("unknown mode: %s", rootOpts.mode)
		}
		switch rootOpts.format {
		case "pretty", "tree", "json", "paths":
		default:
			return cerrs.Newf("unknown format: %s", rootOpts.format)
		}
		switch rootOpts.progress {
		case "on", "off", "auto":
		default:
			return cerrs.Newf("unknown progress value: %s (expected on|off|auto)", rootOpts.progress)
		}
		if rootOpts.algo == "" {
			return cerrs.Newf("algo must not be empty")
		}

		// Validate paths exist and are directories
		if fi, err := os.Stat(left); err != nil {
			return cerrs.Wrapf(err, "left path error: %s", left)
		} else if !fi.IsDir() {
			return cerrs.Newf("left path is not a directory: %s", left)
		}
		if fi, err := os.Stat(right); err != nil {
			return cerrs.Wrapf(err, "right path error: %s", right)
		} else if !fi.IsDir() {
			return cerrs.Newf("right path is not a directory: %s", right)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		left := filepath.Clean(args[0])
		right := filepath.Clean(args[1])

		ui := newProgressUI(shouldEnableProgress(rootOpts.progress))
		ui.Start("Starting…")
		defer ui.Stop()

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
		ui.SetTask("Scanning")
		ui.SetScanDir(left)
		if err := a.ReadFromWithOptions(left, load); err != nil { return cerrs.Wrapf(err, "failed to scan left: %s", left) }
		ui.SetScanDir(right)
		if err := b.ReadFromWithOptions(right, load); err != nil { return cerrs.Wrapf(err, "failed to scan right: %s", right) }

		// Config
		cfg := fsdt.Config{CaseSensitive: !rootOpts.caseInsensitive}
		switch rootOpts.mode {
		case "fast": cfg = fsdt.DefaultFast()
		case "accurate": cfg = fsdt.DefaultAccurate()
		case "checksum": cfg = fsdt.Checksums(rootOpts.algo, store)
		case "checksum-ensure": cfg = fsdt.Checksums(rootOpts.algo, store); cfg.Strategy = fsdt.ChecksumEnsure
		case "checksum-require": cfg = fsdt.ChecksumsStrict(rootOpts.algo, store)
		default:
			return cerrs.Newf("unknown mode: %s", rootOpts.mode)
		}
		cfg.CaseSensitive = !rootOpts.caseInsensitive
		cfg.ExcludeGlobs = append([]string(nil), rootOpts.excludes...)

		// Precompute
		if rootOpts.precompute && store != nil && (cfg.Strategy == fsdt.ChecksumPrefer || cfg.Strategy == fsdt.ChecksumEnsure) {
			leftTotal := countFiles(a)
			rightTotal := countFiles(b)
			ui.SetLeftTotal(leftTotal)
			ui.SetRightTotal(rightTotal)
			ui.SetTask("Precomputing checksums (left)")
			precomputeTreeChecksumsWithProgress(a, rootOpts.algo, store, left, func(){ ui.IncLeftDone() })
			ui.SetTask("Precomputing checksums (right)")
			precomputeTreeChecksumsWithProgress(b, rootOpts.algo, store, right, func(){ ui.IncRightDone() })
		}

		ui.SetTask("Diffing")
		d := fsdt.DiffWithConfig(a, b, cfg)
		if dv, ok := d.Value.(op.DirValue); ok && dv.Reason.Type == op.Because {
			return cerrs.Newf("incompatible or missing prerequisites: %v -> %v", dv.Reason.Before, dv.Reason.After)
		}

		// Stop progress before printing results
		ui.Stop()

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
			return cerrs.Newf("unknown format: %s", rootOpts.format)
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
	rootCmd.Flags().StringVar(&rootOpts.progress, "progress", "auto", "progress: on|off|auto (defaults to auto, renders status to stderr)")
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

func precomputeTreeChecksumsWithProgress(folder *fsdt.Folder, algo string, store fsdt.ChecksumStore, root string, onFile func()) {
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
		for _, inner := range s.Stores { precomputeTreeChecksumsWithProgress(folder, algo, inner, root, onFile) }
		return
	}
	opts.ComputeIfMissing = true
	opts.StreamFromDiskIfAvailable = true
	recursiveEnsureWithProgress(folder, opts, onFile)
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

func recursiveEnsureWithProgress(entry fsdt.FolderEntry, opts fsdt.ChecksumOptions, onFile func()) {
	switch v := entry.(type) {
	case *fsdt.File:
		_, _, _ = v.EnsureChecksum(opts)
		if onFile != nil { onFile() }
	case *fsdt.Folder:
		for _, name := range v.Entries() {
			recursiveEnsureWithProgress(v.Get(name), opts, onFile)
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

func countFiles(entry fsdt.FolderEntry) int {
	switch v := entry.(type) {
	case *fsdt.File:
		return 1
	case *fsdt.Folder:
		total := 0
		for _, name := range v.Entries() {
			total += countFiles(v.Get(name))
		}
		return total
	default:
		return 0
	}
}