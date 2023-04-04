//go:generate mockgen -destination mock_index/mock_index.go github.com/anytypeio/any-sync-filenode/index Index
package index

import (
	"context"
	"errors"
	"github.com/anytypeio/any-sync/app"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
)

var (
	ErrCidsNotExist = errors.New("cids not exist")
)

type Index interface {
	Exists(ctx context.Context, k cid.Cid) (exists bool, err error)
	IsAllExists(ctx context.Context, cids []cid.Cid) (exists bool, err error)
	SpaceInfo(ctx context.Context, spaceId string) (info SpaceInfo, err error)
	GetNonExistentBlocks(ctx context.Context, bs []blocks.Block) (nonExists []blocks.Block, err error)
	Bind(ctx context.Context, spaceId, fileId string, bs []blocks.Block) error
	BindCids(ctx context.Context, spaceId string, fileId string, cids []cid.Cid) error
	UnBind(ctx context.Context, spaceId, fileId string) (err error)
	ExistsInSpace(ctx context.Context, spaceId string, ks []cid.Cid) (exists []cid.Cid, err error)
	FileInfo(ctx context.Context, spaceId, fileId string) (info FileInfo, err error)
	SpaceSize(ctx context.Context, spaceId string) (size uint64, err error)
	Lock(ctx context.Context, ks []cid.Cid) (unlock func(), err error)
	AddBlocks(ctx context.Context, upload []blocks.Block) error
	app.Component
}

type SpaceInfo struct {
	FileCount int
	CidCount  int
}

type FileInfo struct {
	BytesUsage uint64
	CidCount   uint32
}
