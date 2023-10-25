//go:generate mockgen -destination mock_store/mock_store.go github.com/anyproto/any-sync-filenode/store Store
package store

import (
	"context"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/commonfile/fileblockstore"
	"github.com/ipfs/go-cid"
)

type Store interface {
	fileblockstore.BlockStore
	DeleteMany(ctx context.Context, toDelete []cid.Cid) error

	IndexGet(ctx context.Context, key string) (value []byte, err error)
	IndexPut(ctx context.Context, key string, value []byte) (err error)
	app.Component
}
