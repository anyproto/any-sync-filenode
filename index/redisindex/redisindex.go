package redisindex

import (
	"context"
	"github.com/anyproto/any-sync-filenode/index"
	"github.com/anyproto/any-sync-filenode/index/redisindex/indexproto"
	"github.com/anyproto/any-sync-filenode/redisprovider"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
	"time"
)

/*
	Redis db structure:
		CIDS:
			c:{cid}: proto(Entry)

		STORES:
			s:{storeKey}: map
				f:{fileId} -> snappy(proto(CidList))
				{cid} -> int(refCount)
				size -> int(summarySize)

*/

const CName = "filenode.redisindex"

var log = logger.NewNamed(CName)

const (
	storeSizeKey = "size"
)

func New() index.Index {
	return new(redisIndex)
}

type CidInfo struct {
	Cid cid.Cid
	*indexproto.CidEntry
}

type redisIndex struct {
	cl      redis.UniversalClient
	redsync *redsync.Redsync
}

func (r *redisIndex) Init(a *app.App) (err error) {
	r.cl = a.MustComponent(redisprovider.CName).(redisprovider.RedisProvider).Redis()
	r.redsync = redsync.New(goredis.NewPool(r.cl))
	return nil
}

func (r *redisIndex) Name() (name string) {
	return CName
}

func (r *redisIndex) Exists(ctx context.Context, k cid.Cid) (exists bool, err error) {
	res, err := r.cl.Exists(ctx, cidKey(k.String())).Result()
	if err != nil {
		return
	}
	return res != 0, nil
}

func (r *redisIndex) IsAllExists(ctx context.Context, cids []cid.Cid) (exists bool, err error) {
	for _, c := range cids {
		if ex, e := r.Exists(ctx, c); e != nil {
			return false, e
		} else if !ex {
			return false, nil
		}
	}
	return true, nil
}

func (r *redisIndex) GetNonExistentBlocks(ctx context.Context, bs []blocks.Block) (nonExistent []blocks.Block, err error) {
	nonExistent = make([]blocks.Block, 0, len(bs))
	for _, b := range bs {
		if ex, e := r.Exists(ctx, b.Cid()); e != nil {
			return nil, e
		} else if !ex {
			nonExistent = append(nonExistent, b)
		}
	}
	return
}

func (r *redisIndex) Bind(ctx context.Context, key, fileId string, bs []blocks.Block) error {
	bop := bindOp{
		sk: storageKey(key),
		fk: fileIdKey(fileId),
		ri: r,
	}
	newCids, err := bop.Bind(ctx, r.cidInfoByBlocks(bs))
	if err != nil {
		return err
	}
	return r.cidsAddRef(ctx, newCids)
}

func (r *redisIndex) BindCids(ctx context.Context, key, fileId string, ks []cid.Cid) error {
	cids, err := r.cidInfoByKeys(ctx, ks)
	if err != nil {
		return err
	}
	if len(cids) != len(ks) {
		return index.ErrCidsNotExist
	}
	bop := bindOp{
		sk: storageKey(key),
		fk: fileIdKey(fileId),
		ri: r,
	}
	newCids, err := bop.Bind(ctx, cids)
	if err != nil {
		return err
	}
	return r.cidsAddRef(ctx, newCids)
}

func (r *redisIndex) UnBind(ctx context.Context, spaceId, fileId string) (err error) {
	uop := unbindOp{
		sk: storageKey(spaceId),
		fk: fileIdKey(fileId),
		ri: r,
	}
	removedCids, err := uop.Unbind(ctx)
	if err != nil {
		return
	}
	return r.cidsRemoveRef(ctx, removedCids)
}

func (r *redisIndex) ExistsInStorage(ctx context.Context, spaceId string, ks []cid.Cid) (exists []cid.Cid, err error) {
	var sk = storageKey(spaceId)
	cidKeys := make([]string, len(ks))
	for i, k := range ks {
		cidKeys[i] = k.String()
	}
	result, err := r.cl.HMGet(ctx, sk, cidKeys...).Result()
	if err != nil {
		return
	}
	exists = make([]cid.Cid, 0, len(ks))
	for i, v := range result {
		if v != nil {
			exists = append(exists, ks[i])
		}
	}
	return
}

func (r *redisIndex) StorageSize(ctx context.Context, key string) (size uint64, err error) {
	result, err := r.cl.HGet(ctx, storageKey(key), storeSizeKey).Result()
	if err != nil {
		if err == redis.Nil {
			err = nil
		}
		return
	}
	return strconv.ParseUint(result, 10, 64)
}

func (r *redisIndex) StorageInfo(ctx context.Context, key string) (info index.StorageInfo, err error) {
	res, err := r.cl.HKeys(ctx, storageKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			err = nil
		}
		return
	}
	for _, r := range res {
		if strings.HasPrefix(r, "f:") {
			info.FileCount++
		} else if r != storeSizeKey {
			info.CidCount++
		}
	}
	info.Key = key
	return
}

func (r *redisIndex) FileInfo(ctx context.Context, key, fileId string) (info index.FileInfo, err error) {
	fcl, err := r.newFileCidList(ctx, storageKey(key), fileIdKey(fileId))
	if err != nil {
		return
	}
	defer fcl.Unlock()
	cids, err := r.cidInfoByByteKeys(ctx, fcl.Cids)
	if err != nil {
		return
	}
	for _, c := range cids {
		info.CidCount++
		info.BytesUsage += c.Size_
	}
	return
}

func (r *redisIndex) MoveStorage(ctx context.Context, fromKey, toKey string) (err error) {
	ok, err := r.cl.RenameNX(ctx, storageKey(fromKey), storageKey(toKey)).Result()
	if err != nil {
		if err.Error() == "ERR no such key" {
			return index.ErrStorageNotFound
		}
		return err
	}
	if !ok {
		ex, err := r.cl.Exists(ctx, storageKey(toKey)).Result()
		if err != nil {
			return err
		}
		if ex > 0 {
			return index.ErrTargetStorageExists
		}
	}
	return
}

func (r *redisIndex) cidsAddRef(ctx context.Context, cids []CidInfo) error {
	now := time.Now()
	for _, c := range cids {
		ck := cidKey(c.Cid.String())
		res, err := r.cl.Get(ctx, ck).Result()
		if err == redis.Nil {
			err = nil
		}
		if err != nil {
			return err
		}
		entry := &indexproto.CidEntry{}
		if len(res) != 0 {
			if err = entry.Unmarshal([]byte(res)); err != nil {
				return err
			}
		}
		if entry.CreateTime == 0 {
			entry.CreateTime = now.Unix()
		}
		entry.UpdateTime = now.Unix()
		entry.Refs++
		entry.Size_ = c.Size_

		data, _ := entry.Marshal()
		if err = r.cl.Set(ctx, ck, data, 0).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *redisIndex) cidsRemoveRef(ctx context.Context, cids []CidInfo) error {
	now := time.Now()
	for _, c := range cids {
		ck := cidKey(c.Cid.String())
		res, err := r.cl.Get(ctx, ck).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return err
		}
		if len(res) == 0 {
			continue
		}
		entry := &indexproto.CidEntry{}
		if err = entry.Unmarshal([]byte(res)); err != nil {
			return err
		}
		entry.UpdateTime = now.Unix()
		entry.Refs--

		// TODO: syncpool
		data, _ := entry.Marshal()
		if err = r.cl.Set(ctx, ck, data, 0).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *redisIndex) Lock(ctx context.Context, ks []cid.Cid) (unlock func(), err error) {
	var lockers = make([]*redsync.Mutex, 0, len(ks))
	unlock = func() {
		for _, l := range lockers {
			_, _ = l.Unlock()
		}
	}
	for _, k := range ks {
		l := r.redsync.NewMutex("_lock:" + k.String())
		if err = l.LockContext(ctx); err != nil {
			unlock()
			return nil, err
		}
		lockers = append(lockers, l)
	}
	return
}

func (r *redisIndex) AddBlocks(ctx context.Context, bs []blocks.Block) error {
	cids := r.cidInfoByBlocks(bs)
	if err := r.cidsAddRef(ctx, cids); err != nil {
		return err
	}
	return r.cidsRemoveRef(ctx, cids)
}

func (r *redisIndex) cidInfoByBlocks(bs []blocks.Block) (info []CidInfo) {
	info = make([]CidInfo, len(bs))
	for i := range bs {
		info[i] = CidInfo{
			Cid: bs[i].Cid(),
			CidEntry: &indexproto.CidEntry{
				Size_: uint64(len(bs[i].RawData())),
			},
		}
	}
	return
}

func (r *redisIndex) cidInfoByKeys(ctx context.Context, ks []cid.Cid) (info []CidInfo, err error) {
	info = make([]CidInfo, 0, len(ks))
	for _, c := range ks {
		var res string
		res, err = r.cl.Get(ctx, cidKey(c.String())).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return
		}
		entry := &indexproto.CidEntry{}
		if err = entry.Unmarshal([]byte(res)); err != nil {
			return
		}
		info = append(info, CidInfo{
			Cid:      c,
			CidEntry: entry,
		})
	}
	return
}

func (r *redisIndex) cidInfoByByteKeys(ctx context.Context, ks [][]byte) (info []CidInfo, err error) {
	var cids = make([]cid.Cid, 0, len(ks))
	for _, k := range ks {
		if c, e := cid.Cast(k); e == nil {
			cids = append(cids, c)
		}
	}
	return r.cidInfoByKeys(ctx, cids)
}

func storageKey(storeKey string) string {
	return "s:" + storeKey
}

func fileIdKey(fileId string) string {
	return "f:" + fileId
}

func cidKey(k string) string {
	return "c:" + k
}
