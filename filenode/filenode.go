package filenode

import (
	"context"
	"errors"
	"slices"

	"github.com/anyproto/any-sync/acl"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/commonfile/fileblockstore"
	"github.com/anyproto/any-sync/commonfile/fileproto"
	"github.com/anyproto/any-sync/commonfile/fileproto/fileprotoerr"
	"github.com/anyproto/any-sync/metric"
	"github.com/anyproto/any-sync/net/peer"
	"github.com/anyproto/any-sync/net/rpc/server"
	"github.com/anyproto/any-sync/nodeconf"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"go.uber.org/zap"

	"github.com/anyproto/any-sync-filenode/config"
	"github.com/anyproto/any-sync-filenode/index"
	"github.com/anyproto/any-sync-filenode/store"
)

const CName = "filenode.filenode"

var log = logger.NewNamed(CName)

var (
	ErrWrongHash = errors.New("wrong hash")
)

func New() Service {
	return new(fileNode)
}

type Service interface {
	app.Component
}

type fileNode struct {
	acl        acl.AclService
	index      index.Index
	store      store.Store
	metric     metric.Metric
	nodeConf   nodeconf.Service
	migrateKey string
	handler    *rpcHandler
}

func (fn *fileNode) Init(a *app.App) (err error) {
	fn.acl = a.MustComponent(acl.CName).(acl.AclService)
	fn.store = a.MustComponent(fileblockstore.CName).(store.Store)
	fn.index = a.MustComponent(index.CName).(index.Index)
	fn.handler = &rpcHandler{f: fn}
	fn.metric = a.MustComponent(metric.CName).(metric.Metric)
	fn.migrateKey = a.MustComponent(config.CName).(*config.Config).CafeMigrateKey
	fn.nodeConf = a.MustComponent(nodeconf.CName).(nodeconf.Service)
	return fileproto.DRPCRegisterFile(a.MustComponent(server.CName).(server.DRPCServer), fn.handler)
}

func (fn *fileNode) Name() (name string) {
	return CName
}

func (fn *fileNode) Get(ctx context.Context, k cid.Cid, wait bool) (blocks.Block, error) {
	if wait {
		if err := fn.index.WaitCidExists(ctx, k); err != nil {
			return nil, err
		}
	} else {
		exists, err := fn.index.CidExists(ctx, k)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fileprotoerr.ErrCIDNotFound
		}
	}
	return fn.store.Get(ctx, k)
}

func (fn *fileNode) Add(ctx context.Context, spaceId string, fileId string, bs []blocks.Block) error {
	if fileId != "" && fileId == fn.migrateKey {
		return fn.MigrateCafe(ctx, bs)
	}
	storeKey, err := fn.StoreKey(ctx, spaceId, true)
	if err != nil {
		return err
	}
	unlock, err := fn.index.BlocksLock(ctx, bs)
	if err != nil {
		return err
	}
	defer unlock()
	toUpload, err := fn.index.BlocksGetNonExistent(ctx, bs)
	if err != nil {
		return err
	}
	if len(toUpload) > 0 {
		if err = fn.store.Add(ctx, toUpload); err != nil {
			return err
		}
		if err = fn.index.BlocksAdd(ctx, bs); err != nil {
			return err
		}
		fn.index.OnBlockUploaded(ctx, bs...)
	}
	cidEntries, err := fn.index.CidEntriesByBlocks(ctx, bs)
	if err != nil {
		return err
	}
	defer cidEntries.Release()
	return fn.index.FileBind(ctx, storeKey, fileId, cidEntries)
}

func (fn *fileNode) Check(ctx context.Context, spaceId string, cids ...cid.Cid) (result []*fileproto.BlockAvailability, err error) {
	var storeKey index.Key
	if spaceId != "" {
		if storeKey, err = fn.StoreKey(ctx, spaceId, false); err != nil {
			return
		}
	}
	result = make([]*fileproto.BlockAvailability, 0, len(cids))
	var inSpaceM = make(map[string]struct{})
	if spaceId != "" {
		inSpace, err := fn.index.CidExistsInSpace(ctx, storeKey, cids)
		if err != nil {
			return nil, err
		}
		for _, inSp := range inSpace {
			inSpaceM[inSp.KeyString()] = struct{}{}
		}
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
			if ex, err = fn.index.CidExists(ctx, k); err != nil {
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
	storeKey, err := fn.StoreKey(ctx, spaceId, true)
	if err != nil {
		return err
	}
	cidEntries, err := fn.index.CidEntries(ctx, cids)
	if err != nil {
		return err
	}
	defer cidEntries.Release()
	return fn.index.FileBind(ctx, storeKey, fileId, cidEntries)
}

func (fn *fileNode) StoreKey(ctx context.Context, spaceId string, checkLimit bool) (storageKey index.Key, err error) {
	if spaceId == "" {
		return storageKey, fileprotoerr.ErrForbidden
	}

	identity, err := peer.CtxPubKey(ctx)
	if err != nil {
		return storageKey, fileprotoerr.ErrForbidden
	}

	ownerPubKey, err := fn.acl.OwnerPubKey(ctx, spaceId)
	if err != nil {
		log.WarnCtx(ctx, "acl ownerPubKey error", zap.Error(err))
		return storageKey, fileprotoerr.ErrForbidden
	}
	storageKey = index.Key{
		GroupId: ownerPubKey.Account(),
		SpaceId: spaceId,
	}

	// if it not owner
	if identity.Account() != storageKey.GroupId {
		permissions, err := fn.acl.Permissions(ctx, identity, spaceId)
		if err != nil {
			log.WarnCtx(ctx, "acl permissions error", zap.Error(err))
			return storageKey, fileprotoerr.ErrForbidden
		}
		if !permissions.CanWrite() {
			return storageKey, fileprotoerr.ErrForbidden
		}
	}

	if e := fn.index.Migrate(ctx, storageKey); e != nil {
		log.WarnCtx(ctx, "space migrate error", zap.String("spaceId", spaceId), zap.Error(e))
	}

	if checkLimit {
		if err = fn.index.CheckLimits(ctx, storageKey); err != nil {
			if errors.Is(err, index.ErrLimitExceed) {
				return storageKey, fileprotoerr.ErrSpaceLimitExceeded
			} else {
				log.WarnCtx(ctx, "check limit error", zap.Error(err))
				return storageKey, fileprotoerr.ErrUnexpected
			}
		}
	}
	return
}

func (fn *fileNode) SpaceInfo(ctx context.Context, spaceId string) (info *fileproto.SpaceInfoResponse, err error) {
	storageKey, err := fn.StoreKey(ctx, spaceId, false)
	if err != nil {
		return nil, err
	}
	groupInfo, err := fn.index.GroupInfo(ctx, storageKey.GroupId)
	if err != nil {
		return nil, err
	}
	if info, err = fn.spaceInfo(ctx, storageKey, groupInfo); err != nil {
		return nil, err
	}
	return
}

func (fn *fileNode) AccountInfo(ctx context.Context) (info *fileproto.AccountInfoResponse, err error) {
	info = &fileproto.AccountInfoResponse{}

	identity, err := peer.CtxPubKey(ctx)
	if err != nil {
		return nil, fileprotoerr.ErrForbidden
	}

	groupId := identity.Account()

	groupInfo, err := fn.index.GroupInfo(ctx, groupId)
	if err != nil {
		return nil, err
	}
	info.TotalCidsCount = groupInfo.CidsCount
	info.TotalUsageBytes = groupInfo.BytesUsage
	info.LimitBytes = groupInfo.Limit
	info.AccountLimitBytes = groupInfo.AccountLimit
	for _, spaceId := range groupInfo.SpaceIds {
		spaceInfo, err := fn.spaceInfo(ctx, index.Key{GroupId: groupId, SpaceId: spaceId}, groupInfo)
		if err != nil {
			return nil, err
		}
		info.Spaces = append(info.Spaces, spaceInfo)
	}
	return
}

func (fn *fileNode) spaceInfo(ctx context.Context, key index.Key, groupInfo index.GroupInfo) (info *fileproto.SpaceInfoResponse, err error) {
	info = &fileproto.SpaceInfoResponse{}
	info.SpaceId = key.SpaceId
	spaceInfo, err := fn.index.SpaceInfo(ctx, key)
	if err != nil {
		return nil, err
	}
	if spaceInfo.Limit == 0 {
		info.TotalUsageBytes = groupInfo.BytesUsage
		info.LimitBytes = groupInfo.Limit
	} else {
		info.TotalUsageBytes = spaceInfo.BytesUsage
		info.LimitBytes = spaceInfo.Limit
	}
	info.FilesCount = uint64(spaceInfo.FileCount)
	info.CidsCount = spaceInfo.CidsCount
	info.SpaceUsageBytes = spaceInfo.BytesUsage
	return
}

func (fn *fileNode) FilesDelete(ctx context.Context, spaceId string, fileIds []string) (err error) {
	storeKey, err := fn.StoreKey(ctx, spaceId, false)
	if err != nil {
		return
	}
	return fn.index.FileUnbind(ctx, storeKey, fileIds...)
}

func (fn *fileNode) FileInfo(ctx context.Context, spaceId string, fileIds ...string) (info []*fileproto.FileInfo, err error) {
	storeKey, err := fn.StoreKey(ctx, spaceId, false)
	if err != nil {
		return
	}
	fis, err := fn.index.FileInfo(ctx, storeKey, fileIds...)
	if err != nil {
		return nil, err
	}
	info = make([]*fileproto.FileInfo, len(fis))
	for i, fi := range fis {
		info[i] = &fileproto.FileInfo{
			FileId:     fileIds[i],
			UsageBytes: fi.BytesUsage,
			CidsCount:  uint32(fi.CidsCount),
		}
	}
	return
}

func (fn *fileNode) MigrateCafe(ctx context.Context, bs []blocks.Block) error {
	unlock, err := fn.index.BlocksLock(ctx, bs)
	if err != nil {
		return err
	}
	defer unlock()

	if err = fn.store.Add(ctx, bs); err != nil {
		return err
	}
	if err = fn.index.BlocksAdd(ctx, bs); err != nil {
		return err
	}
	return nil
}

func (fn *fileNode) AccountLimitSet(ctx context.Context, identity string, limit uint64) (err error) {
	peerId, err := peer.CtxPeerId(ctx)
	if err != nil {
		return
	}
	// check that call from the coordinator or the payment node
	if !slices.Contains(fn.nodeConf.NodeTypes(peerId), nodeconf.NodeTypeCoordinator) &&
		!slices.Contains(fn.nodeConf.NodeTypes(peerId), nodeconf.NodeTypePaymentProcessingNode) {
		return fileprotoerr.ErrForbidden
	}

	return fn.index.SetGroupLimit(ctx, identity, limit)
}

func (fn *fileNode) SpaceLimitSet(ctx context.Context, spaceId string, limit uint64) (err error) {
	storeKey, err := fn.StoreKey(ctx, spaceId, false)
	if err != nil {
		return
	}
	return fn.index.SetSpaceLimit(ctx, storeKey, limit)
}

func (fn *fileNode) FilesGet(ctx context.Context, spaceId string) (fileIds []string, err error) {
	storeKey, err := fn.StoreKey(ctx, spaceId, false)
	if err != nil {
		return
	}
	return fn.index.FilesList(ctx, storeKey)
}
