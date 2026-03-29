package store

import (
	"context"
	"io"
)

// ObjectStore stores and retrieves objects by bucket and key.
//
// Bucket and key must be non empty for now
// The reader returned by Get must be called by the caller.
//
// Current implementation do accept a context.Context but makes no use of that
type ObjectStore interface {
	// Stores the contents of r at bucket/key
	// Replaces the objext if it already exists
	//FIX: does partial writes for now
	Put(ctx context.Context, bucket, key string, r io.Reader) error

	// Returns a reader for the object at bucket/key
	// The caller must call the Close()
	Get(ctx context.Context, bucket, key string) (io.ReadCloser, error)

	// Removes the object at bucket/key
	Delete(ctx context.Context, bucket, key string) error
}
