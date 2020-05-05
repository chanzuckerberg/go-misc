package storage

import "context"

const (
	service = "aws-oidc"

	storageVersion = "v0"
)

// Storage represents a storage backend for a cache
type Storage interface {
	Read(context.Context) (*string, error)
	Set(ctx context.Context, value string) error
	Delete(context.Context) error
}
