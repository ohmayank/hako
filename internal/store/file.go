package store

import (
	"errors"
	"path/filepath"
	"strings"
)

var (
	ErrWrongPath   = errors.New("The root and final path are not the same")
	ErrEmptyBucket = errors.New("The bucket has to be non empty")
	ErrEmptyKey    = errors.New("The key has to be non empty")
)

type FileStore struct {
	root string
}

func newFileStore(r string) *FileStore {
	return &FileStore{root: r}
}

// Constructs the path.
// Both bucket and key must explicitly be non empty.
func (fs *FileStore) path(bucket, key string) (string, error) {
	if bucket == "" {
		return "", ErrEmptyBucket
	}
	if key == "" {
		return "", ErrEmptyKey
	}

	root := filepath.Clean(fs.root)
	p := filepath.Join(root, bucket, key)

	rel, err := filepath.Rel(fs.root, p)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", ErrWrongPath
	}

	return p, nil

}
