package serverstore

import (
	"context"
	"github.com/anytypeio/any-sync/commonfile/fileblockstore"
	"github.com/anytypeio/any-sync/commonfile/fileproto"
	"github.com/ipfs/go-cid"
)

type ServerStore interface {
	fileblockstore.BlockStore
	Check(ctx context.Context, spaceId string, cids ...cid.Cid) (result []*fileproto.BlockAvailability, err error)
}
