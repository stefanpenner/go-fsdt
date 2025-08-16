# go-fsdt

Fast, configurable filesystem diffing for Go.

> Alpha: APIs may change.

### Features
- **Diff directories**: compute minimal operations between two trees
- **Modes**: fast, accurate, checksum (+ ensure/require)
- **Checksum stores**: xattr or sidecar cache
- **Globs**: doublestar excludes
- **Pretty/JSON/paths** output

### Install
- CLI: `go install github.com/stefanpenner/go-fsdt/cmd/fsdt@latest`
- Library: `go get github.com/stefanpenner/go-fsdt@latest`

### CLI
- Usage: `fsdt [flags] <left> <right>`
- Common flags:
  - `--mode` fast|accurate|checksum|checksum-ensure|checksum-require
  - `--algo` sha256 (for checksum modes)
  - `--xattr` key (e.g. `user.sha256` on Linux, `com.yourorg.sha256` on macOS)
  - `--sidecar` DIR (alias: `--checksum-cache-dir`), `--root` PATH, `--precompute`
  - `--ci` case-insensitive, `--exclude` GLOB (repeat), `--format` pretty|tree|json|paths

Example:
```bash
fsdt --mode accurate --format tree --exclude "**/.git/**" ./left ./right
```

### Library (tiny example)
```go
import (
  fsdt "github.com/stefanpenner/go-fsdt"
  op "github.com/stefanpenner/go-fsdt/operation"
)

a := fsdt.NewFolder()
b := fsdt.NewFolder()
a.FileString("README.md", "hello\n")

cfg := fsdt.DefaultAccurate()
d := fsdt.DiffWithConfig(a, b, cfg)
_ = op.Print(d) // pretty string
```

### Contributing
Issues and PRs are welcome.

### License
[MIT](LICENSE)
