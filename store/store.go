//go:generate mockgen -destination mock_store/mock_store.go github.com/anytypeio/any-sync-filenode/store Store
package store

import (
	"context"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/commonfile/fileblockstore"
	"github.com/ipfs/go-cid"
)

type Store interface {
	fileblockstore.BlockStore
	DeleteMany(ctx context.Context, toDelete []cid.Cid) error
	app.Component
}
