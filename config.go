package fsdt

type CompareStrategy int

const (
	StructureOnly CompareStrategy = iota
	Bytes
	ChecksumPrefer
	ChecksumRequire
)

type Config struct {
	// Compare
	CaseSensitive bool
	CompareMode   bool
	CompareSize   bool
	CompareMTime  bool
	Strategy      CompareStrategy
	ExcludeGlobs  []string

	// Cache
	Algorithm string
	Store     ChecksumStore
}

func DefaultFast() Config {
	return Config{
		CaseSensitive: true,
		CompareMode:   true,
		Strategy:      StructureOnly,
	}
}

func DefaultAccurate() Config {
	return Config{
		CaseSensitive: true,
		CompareMode:   true,
		CompareSize:   true,
		CompareMTime:  true,
		Strategy:      Bytes,
	}
}

func Checksums(algo string, store ChecksumStore) Config {
	return Config{
		CaseSensitive: true,
		CompareMode:   true,
		Strategy:      ChecksumPrefer,
		Algorithm:     algo,
		Store:         store,
	}
}

func ChecksumsStrict(algo string, store ChecksumStore) Config {
	cfg := Checksums(algo, store)
	cfg.Strategy = ChecksumRequire
	return cfg
}