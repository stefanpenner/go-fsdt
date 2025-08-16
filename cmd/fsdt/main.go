package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	fsdt "github.com/stefanpenner/go-fsdt"
	op "github.com/stefanpenner/go-fsdt/operation"
)

type stringSliceFlag []string

func (s *stringSliceFlag) String() string { return strings.Join(*s, ",") }
func (s *stringSliceFlag) Set(v string) error {
	if v == "" { return nil }
	*s = append(*s, v)
	return nil
}

func main() {
	var (
		mode      = flag.String("mode", "accurate", "diff mode: fast|accurate|checksum|checksum-strict")
		algo      = flag.String("algo", "sha256", "checksum algorithm when using checksum modes")
		xattrKey  = flag.String("xattr", "", "xattr key for reading/writing checksums (e.g., user.sha256)")
		sidecar   = flag.String("sidecar", "", "sidecar directory to store checksums")
		root      = flag.String("root", "", "common root for sidecar relative paths (defaults to left path)")
		precompute = flag.Bool("precompute", false, "precompute and persist checksums before diff (if store configured)")
		caseInsensitive = flag.Bool("ci", false, "case-insensitive diff")
		format    = flag.String("format", "pretty", "output format: pretty|tree|json|paths")
	)
	var excludes stringSliceFlag
	flag.Var(&excludes, "exclude", "exclude glob (repeatable)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <left> <right>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}
	left := filepath.Clean(flag.Arg(0))
	right := filepath.Clean(flag.Arg(1))
	if *root == "" { *root = left }

	// Build store from flags
	var stores []fsdt.ChecksumStore
	if *xattrKey != "" {
		stores = append(stores, fsdt.XAttrStore{Key: *xattrKey})
	}
	if *sidecar != "" {
		stores = append(stores, fsdt.SidecarStore{BaseDir: *sidecar, Root: *root, Algorithm: *algo})
	}
	var store fsdt.ChecksumStore
	if len(stores) == 1 { store = stores[0] }
	if len(stores) > 1 { store = fsdt.MultiStore{Stores: stores} }

	// Load trees
	a := fsdt.NewFolder()
	b := fsdt.NewFolder()
	load := fsdt.LoadOptions{}
	if *xattrKey != "" {
		load.XAttrChecksumKey = *xattrKey
		load.ChecksumAlgorithm = *algo
		load.ComputeChecksumIfMissing = false
		load.WriteComputedChecksumToXAttr = false
	}
	if err := a.ReadFromWithOptions(left, load); err != nil { fatal(err) }
	if err := b.ReadFromWithOptions(right, load); err != nil { fatal(err) }

	// Build config
	cfg := fsdt.Config{CaseSensitive: !*caseInsensitive}
	switch *mode {
	case "fast": cfg = fsdt.DefaultFast()
	case "accurate": cfg = fsdt.DefaultAccurate()
	case "checksum": cfg = fsdt.Checksums(*algo, store)
	case "checksum-strict": cfg = fsdt.ChecksumsStrict(*algo, store)
	default:
		fatal(fmt.Errorf("unknown mode: %s", *mode))
	}
	cfg.CaseSensitive = !*caseInsensitive
	cfg.ExcludeGlobs = append([]string(nil), excludes...)

	// Precompute if requested
	if *precompute && store != nil && (cfg.Strategy == fsdt.ChecksumPrefer || cfg.Strategy == fsdt.ChecksumRequire) {
		precomputeTreeChecksums(a, *algo, store, left)
		precomputeTreeChecksums(b, *algo, store, right)
	}

	d := fsdt.DiffWithConfig(a, b, cfg)

	// If the diff indicates incompatible excludes
	if dv, ok := d.Value.(op.DirValue); ok && dv.Reason.Type == op.Because {
		fatal(fmt.Errorf("incompatible exclude globs: left=%v right=%v", dv.Reason.Before, dv.Reason.After))
	}

	switch *format {
	case "pretty":
		fmt.Println(op.Print(d))
	case "tree":
		fmt.Println(op.Print(d))
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(d)
	case "paths":
		for _, p := range collectPaths(d) { fmt.Println(p) }
	default:
		fatal(fmt.Errorf("unknown format: %s", *format))
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}

func precomputeTreeChecksums(folder *fsdt.Folder, algo string, store fsdt.ChecksumStore, root string) {
	opts := fsdt.ChecksumOptions{Algorithm: algo}
	// Map store into options
	switch s := store.(type) {
	case fsdt.XAttrStore:
		opts.XAttrKey = s.Key
		opts.WriteToXAttr = true
	case fsdt.SidecarStore:
		opts.SidecarDir = s.BaseDir
		opts.RootPath = s.Root
		opts.WriteToXAttr = false
	case fsdt.MultiStore:
		// Try to set both if present
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