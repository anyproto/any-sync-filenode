package index

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-libipfs/blocks"
)

type Index interface {
	Exists(ctx context.Context, k cid.Cid) (exists bool, err error)
	FilterExistingOnly(ctx context.Context, cids []cid.Cid) (exists []cid.Cid, err error)
	GetNonExistentBlocks(ctx context.Context, bs []blocks.Block) (nonExists []blocks.Block, err error)
	Bind(ctx context.Context, spaceId string, bs []blocks.Block) error
	UnBind(ctx context.Context, spaceId string, ks []cid.Cid) (toDelete []cid.Cid, err error)
	ExistsInSpace(ctx context.Context, spaceId string, ks []cid.Cid) (exists []cid.Cid, err error)
}
