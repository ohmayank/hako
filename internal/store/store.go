package store

import (
	"context"
	"io"
)

type ObjectStore interface {
	Put(ctx context.Context, bucket, key string, r io.Reader) error
	Get(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string) error
}
