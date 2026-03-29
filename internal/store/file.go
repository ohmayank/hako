package store

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrWrongPath   = errors.New("invalid path")
	ErrEmptyBucket = errors.New("empty bucket")
	ErrEmptyKey    = errors.New("empty key")
)

type FileStore struct {
	root string
}

func NewFileStore(r string) *FileStore {
	return &FileStore{root: r}
}

// Constructs the path.
// Both bucket and key must explicitly be non empty.
//
// TODO: Has some weird key semantics for now
func (fs *FileStore) path(bucket, key string) (string, error) {
	if bucket == "" {
		return "", ErrEmptyBucket
	}
	if key == "" {
		return "", ErrEmptyKey
	}

	root := filepath.Clean(fs.root)
	p := filepath.Join(root, bucket, key)

	rel, err := filepath.Rel(root, p)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", ErrWrongPath
	}

	return p, nil

}

func (fs *FileStore) Put(ctx context.Context, bucket, key string, r io.Reader) error {
	_ = ctx //TODO: no use of ctx for now also fix later
	path, err := fs.path(bucket, key)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	//FIX: later, since io.Copy can technically fail and do a partial write
	if _, err := io.Copy(f, r); err != nil {
		return err
	}

	return nil

}

func (fs *FileStore) Get(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	_ = ctx //TODO: no use of ctx for now also fix later

	path, err := fs.path(bucket, key)
	if err != nil {
		return nil, err
	}

	return os.Open(path)
}

func (fs *FileStore) Delete(ctx context.Context, bucket, key string) error {
	_ = ctx //TODO: no use of ctx for now also fix later

	path, err := fs.path(bucket, key)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
