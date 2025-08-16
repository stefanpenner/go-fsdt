package fsdt

type ChecksumStrategy int

const (
	ChecksumOff ChecksumStrategy = iota
	ChecksumPreferPolicy
	ChecksumEnsurePolicy
	ChecksumRequirePolicy
)

type ChecksumPolicy struct {
	Strategy  ChecksumStrategy
	Algorithm string
	Store     ChecksumStore
	Persist   bool
}

// WithChecksumPrefer sets a policy that prefers checksums (compute if missing) with optional persistence via store.
func WithChecksumPrefer(algo string, store ChecksumStore, persist bool) func(*Folder) {
	return func(f *Folder) {
		f.policy = ChecksumPolicy{Strategy: ChecksumPreferPolicy, Algorithm: algo, Store: store, Persist: persist}
	}
}

// WithChecksumEnsure sets a policy that ensures checksums exist (compute if missing) with optional persistence.
func WithChecksumEnsure(algo string, store ChecksumStore, persist bool) func(*Folder) {
	return func(f *Folder) {
		f.policy = ChecksumPolicy{Strategy: ChecksumEnsurePolicy, Algorithm: algo, Store: store, Persist: persist}
	}
}

// WithChecksumRequire sets a policy that requires checksums exist; typically used with a store.
func WithChecksumRequire(algo string, store ChecksumStore, persist bool) func(*Folder) {
	return func(f *Folder) {
		f.policy = ChecksumPolicy{Strategy: ChecksumRequirePolicy, Algorithm: algo, Store: store, Persist: persist}
	}
}