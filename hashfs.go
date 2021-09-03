// Package hashfs provides a hashing fs.FS implementation.
package hashfs

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"path/filepath"
	"sync"
)

// FS is a fs.FS implementation that appends
// sha256 digests to the filenames.
type FS struct {
	mu   sync.RWMutex
	fs   fs.FS
	hash map[string]string // ["base.ext"] => "hash"
	base map[string]string // ["base.hash.ext"] => "base.ext"
}

// New returns a new hashing fs.FS implementation.
func New(fs fs.FS) *FS {
	return &FS{
		fs:   fs,
		hash: make(map[string]string),
		base: make(map[string]string),
	}
}

// Hash returns the sha256 digest of the given file.
func (f *FS) Hash(name string) string {
	hash, ok := f.getHash(name)
	if ok {
		return hash
	}
	ext := filepath.Ext(name)
	hash = f.makeHash(name)
	if hash == "" {
		return ""
	}
	base := name[:len(name)-len(ext)] + "." + hash + ext
	f.mu.Lock()
	f.hash[name] = hash
	f.base[base] = name
	f.mu.Unlock()
	return hash
}

// getHash performs a synchronized lookup on the hash map.
func (f *FS) getHash(name string) (string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	hash, ok := f.hash[name]
	return hash, ok
}

// makeHash returns the full sha256 digest for the given file name.
func (f *FS) makeHash(name string) string {
	b, err := fs.ReadFile(f.fs, name)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(b)
	return hex.EncodeToString(digest[:])
}

// Name returns the hashed file name for the given file.
func (f *FS) Name(name string) string {
	ext := filepath.Ext(name)
	hash := f.Hash(name)
	if hash == "" {
		return ""
	}
	return name[:len(name)-len(ext)] + "." + hash + ext
}

// Open implements the fs.FS interface.
func (f *FS) Open(name string) (fs.File, error) {
	base, ok := f.getBase(name)
	if ok {
		return f.fs.Open(base)
	}
	ext := filepath.Ext(name)
	if ext == "" {
		// Needs at least one extension to be a request for a hashed file.
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	hashExt := filepath.Ext(name[:len(name)-len(ext)])
	if hashExt == "" {
		// Maybe the only "extension" is the hash itself.
		base = name[:len(name)-len(ext)]
		hashExt = ext
	} else {
		base = name[:len(name)-len(hashExt)-len(ext)] + ext
	}
	hash := f.makeHash(base)
	if hash == "" || hashExt[1:] != hash {
		// Needs to exist and have valid hash.
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	f.mu.Lock()
	f.hash[base] = hash
	f.base[name] = base
	f.mu.Unlock()
	return f.fs.Open(base)
}

// getBase performs a synchronized lookup on the base map.
func (f *FS) getBase(name string) (string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	base, ok := f.base[name]
	return base, ok
}
