package filenode

import (
	"context"
	"errors"
	"fmt"
	"github.com/anytypeio/any-sync-filenode/index"
	"github.com/anytypeio/any-sync-filenode/index/redisindex"
	"github.com/anytypeio/any-sync-filenode/limit"
	"github.com/anytypeio/any-sync-filenode/store"
	"github.com/anytypeio/any-sync-filenode/testutil"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/app/logger"
	"github.com/anytypeio/any-sync/commonfile/fileblockstore"
	"github.com/anytypeio/any-sync/commonfile/fileproto"
	"github.com/anytypeio/any-sync/net/rpc/server"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-libipfs/blocks"
)

const CName = "filenode.filenode"

var log = logger.NewNamed(CName)

var (
	ErrCidsNotExists = errors.New("cids not exists")
	ErrLimitExceed   = errors.New("space limit exceed")

	ErrCidDataTooBig = errors.New("cid data too big")
	ErrWrongHash     = errors.New("wrong hash")
)

func New() Service {
	return new(fileNode)
}

type Service interface {
	app.Component
}

type fileNode struct {
	index   index.Index
	store   store.Store
	limit   limit.Limit
	handler *rpcHandler
}

func (fn *fileNode) Init(a *app.App) (err error) {
	fn.store = a.MustComponent(fileblockstore.CName).(store.Store)
	fn.index = a.MustComponent(redisindex.CName).(index.Index)
	fn.limit = a.MustComponent(limit.CName).(limit.Limit)
	fn.handler = &rpcHandler{f: fn}
	return fileproto.DRPCRegisterFile(a.MustComponent(server.CName).(server.DRPCServer), fn.handler)
}

func (fn *fileNode) Name() (name string) {
	return CName
}

func (fn *fileNode) Get(ctx context.Context, k cid.Cid) (blocks.Block, error) {
	exists, err := fn.index.Exists(ctx, k)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fileblockstore.ErrCIDNotFound
	}
	return fn.store.Get(ctx, k)
}

func (fn *fileNode) Add(ctx context.Context, spaceId string, fileId string, bs []blocks.Block) error {
	if err := fn.ValidateSpaceId(ctx, spaceId, true); err != nil {
		return err
	}
	unlock, err := fn.index.Lock(ctx, testutil.BlocksToKeys(bs))
	if err != nil {
		return err
	}
	defer unlock()
	toUpload, err := fn.index.GetNonExistentBlocks(ctx, bs)
	if err != nil {
		return err
	}
	if len(toUpload) > 0 {
		if err = fn.store.Add(ctx, toUpload); err != nil {
			return err
		}
	}
	return fn.index.Bind(ctx, spaceId, fileId, bs)
}

func (fn *fileNode) Check(ctx context.Context, spaceId string, cids ...cid.Cid) (result []*fileproto.BlockAvailability, err error) {
	if err = fn.ValidateSpaceId(ctx, spaceId, false); err != nil {
		return
	}
	result = make([]*fileproto.BlockAvailability, 0, len(cids))
	inSpace, err := fn.index.ExistsInSpace(ctx, spaceId, cids)
	if err != nil {
		return nil, err
	}
	var inSpaceM = make(map[string]struct{})
	for _, inSp := range inSpace {
		inSpaceM[inSp.KeyString()] = struct{}{}
	}

	for _, k := range cids {
		var res = &fileproto.BlockAvailability{
			Cid:    k.Bytes(),
			Status: fileproto.AvailabilityStatus_NotExists,
		}
		if _, ok := inSpaceM[k.KeyString()]; ok {
			res.Status = fileproto.AvailabilityStatus_ExistsInSpace
		} else {
			var ex bool
			if ex, err = fn.index.Exists(ctx, k); err != nil {
				return nil, err
			} else if ex {
				res.Status = fileproto.AvailabilityStatus_Exists
			}
		}
		result = append(result, res)
	}
	return
}

func (fn *fileNode) BlocksBind(ctx context.Context, spaceId, fileId string, cids ...cid.Cid) (err error) {
	if err = fn.ValidateSpaceId(ctx, spaceId, true); err != nil {
		return err
	}
	unlock, err := fn.index.Lock(ctx, cids)
	if err != nil {
		return err
	}
	defer unlock()
	return fn.index.BindCids(ctx, spaceId, fileId, cids)
}

func (fn *fileNode) ValidateSpaceId(ctx context.Context, spaceId string, checkLimit bool) (err error) {
	if spaceId == "" {
		return fmt.Errorf("empty space id")
	}
	// this call also confirms that space exists and valid
	limitBytes, err := fn.limit.Check(ctx, spaceId)
	if err != nil {
		return
	}
	if checkLimit {
		currentSize, e := fn.index.SpaceSize(ctx, spaceId)
		if e != nil {
			return e
		}
		if currentSize >= limitBytes {
			return ErrLimitExceed
		}
	}
	return
}

func (fn *fileNode) SpaceInfo(ctx context.Context, spaceId string) (info *fileproto.SpaceInfoResponse, err error) {
	info = &fileproto.SpaceInfoResponse{}
	// we have space/identity validation in limit.Check
	if info.LimitBytes, err = fn.limit.Check(ctx, spaceId); err != nil {
		return nil, err
	}
	if info.UsageBytes, err = fn.index.SpaceSize(ctx, spaceId); err != nil {
		return nil, err
	}
	si, err := fn.index.SpaceInfo(ctx, spaceId)
	if err != nil {
		return nil, err
	}
	info.CidsCount = uint64(si.CidCount)
	info.FilesCount = uint64(si.FileCount)
	return
}

func (fn *fileNode) FilesDelete(ctx context.Context, spaceId string, fileIds []string) (err error) {
	if err = fn.ValidateSpaceId(ctx, spaceId, false); err != nil {
		return
	}
	for _, fileId := range fileIds {
		if err = fn.index.UnBind(ctx, spaceId, fileId); err != nil {
			return
		}
	}
	return
}

func (fn *fileNode) FileInfo(ctx context.Context, spaceId, fileId string) (info *fileproto.FileInfo, err error) {
	if err = fn.ValidateSpaceId(ctx, spaceId, false); err != nil {
		return
	}
	fi, err := fn.index.FileInfo(ctx, spaceId, fileId)
	if err != nil {
		return nil, err
	}
	return &fileproto.FileInfo{
		FileId:     fileId,
		UsageBytes: fi.BytesUsage,
		CidsCount:  fi.CidCount,
	}, nil
}
