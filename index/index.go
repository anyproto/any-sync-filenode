//go:generate mockgen -destination mock_index/mock_index.go github.com/anyproto/any-sync-filenode/index Index
package index

import (
	"context"
	"errors"
	"github.com/anyproto/any-sync/app"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
)

var (
	ErrCidsNotExist        = errors.New("cids not exist")
	ErrTargetStorageExists = errors.New("target storage exists")
	ErrStorageNotFound     = errors.New("storage not found")
)

type Index interface {
	Exists(ctx context.Context, k cid.Cid) (exists bool, err error)
	IsAllExists(ctx context.Context, cids []cid.Cid) (exists bool, err error)
	StorageInfo(ctx context.Context, key string) (info StorageInfo, err error)
	GetNonExistentBlocks(ctx context.Context, bs []blocks.Block) (nonExists []blocks.Block, err error)
	Bind(ctx context.Context, key, fileId string, bs []blocks.Block) error
	BindCids(ctx context.Context, key string, fileId string, cids []cid.Cid) error
	UnBind(ctx context.Context, key, fileId string) (err error)
	ExistsInStorage(ctx context.Context, key string, ks []cid.Cid) (exists []cid.Cid, err error)
	FileInfo(ctx context.Context, key, fileId string) (info FileInfo, err error)
	StorageSize(ctx context.Context, key string) (size uint64, err error)
	Lock(ctx context.Context, ks []cid.Cid) (unlock func(), err error)
	AddBlocks(ctx context.Context, upload []blocks.Block) error
	MoveStorage(ctx context.Context, fromKey, toKey string) error
	app.Component
}

type StorageInfo struct {
	Key       string
	FileCount int
	CidCount  int
}

type FileInfo struct {
	BytesUsage uint64
	CidCount   uint32
}
