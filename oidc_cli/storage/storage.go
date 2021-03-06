package storage

import (
	"context"

	"github.com/chanzuckerberg/go-misc/oidc_cli/client"
)

const (
	service = "aws-oidc"

	storageVersion = "v0"
)

// Storage represents a storage backend for a cache
type Storage interface {
	Read(context.Context) (*string, error)
	Set(ctx context.Context, value string) error
	Delete(context.Context) error

	MarshalOpts() []client.MarshalOpts
}
