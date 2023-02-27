package redisindex

import (
	"context"
	"github.com/anytypeio/any-sync-filenode/index"
	"github.com/anytypeio/any-sync-filenode/redisprovider"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/app/logger"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-libipfs/blocks"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

const CName = "filenode.redisindex"

var log = logger.NewNamed(CName)

const (
	spaceSizeKey = "size"
)

func New() index.Index {
	return new(redisIndex)
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
	res, err := r.cl.Get(ctx, cidKey(k.String())).Result()
	if err != nil {
		if err == redis.Nil {
			err = nil
		}
		return
	}
	return res != "0", nil
}

func (r *redisIndex) FilterExistingOnly(ctx context.Context, cids []cid.Cid) (exists []cid.Cid, err error) {
	// we can't use batch operations with cids because we use redis-cluster
	exists = cids[:0]
	for _, c := range cids {
		if ex, e := r.Exists(ctx, c); e != nil {
			return nil, e
		} else if ex {
			exists = append(exists, c)
		}
	}
	return
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

func (r *redisIndex) Bind(ctx context.Context, spaceId string, bs []blocks.Block) error {
	now := time.Now()
	var sk = spaceKey(spaceId)
	bs = uniqBlocks(bs)
	cidKeys := make([]string, len(bs))
	toBindKV := make([]interface{}, 0, len(bs)*2)
	for i, b := range bs {
		cidKeys[i] = b.Cid().String()
		size := uint64(len(b.RawData()))
		toBindKV = append(toBindKV, cidKeys[i], Entry{
			Size: size,
			Time: now,
		}.Binary())
	}
	var (
		getRes *redis.SliceCmd
		err    error
	)

	_, err = r.cl.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		getRes = pipe.HMGet(ctx, sk, cidKeys...)
		if err = getRes.Err(); err != nil {
			return err
		}
		return pipe.HMSet(ctx, sk, toBindKV...).Err()
	})
	if err != nil {
		return err
	}
	var getVals = getRes.Val()
	var addedSize uint64
	var toBind = bs[:0]
	for i, b := range bs {
		if getVals[i] == nil { // nil means this cid didn't exist before we added it
			addedSize += uint64(len(b.RawData()))
			toBind = append(toBind, b)
		}
	}

	if len(toBind) == 0 {
		return nil
	}

	if err = r.cl.HIncrBy(ctx, sk, spaceSizeKey, int64(addedSize)).Err(); err != nil {
		return err
	}

	// increment ref counter for cids
	for _, b := range toBind {
		if err = r.cl.Incr(ctx, cidKey(b.Cid().String())).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *redisIndex) UnBind(ctx context.Context, spaceId string, ks []cid.Cid) (toDelete []cid.Cid, err error) {
	ks = uniqKeys(ks)
	var sk = spaceKey(spaceId)
	cidKeys := make([]string, len(ks))
	for i, k := range ks {
		cidKeys[i] = k.String()
	}

	var getRes *redis.SliceCmd

	_, err = r.cl.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		getRes = pipe.HMGet(ctx, sk, cidKeys...)
		if err = getRes.Err(); err != nil {
			return err
		}
		return pipe.HDel(ctx, sk, cidKeys...).Err()
	})
	if err != nil {
		return nil, err
	}

	var deletedSize uint64
	var existingKeys = cidKeys[:0]
	var getVals = getRes.Val()
	for i, k := range cidKeys {
		if v := getVals[i]; v != nil {
			if b, ok := v.(string); ok {
				if en, e := NewEntry([]byte(b)); e == nil {
					deletedSize += en.Size
				} else {
					panic(1)
				}
			}
			existingKeys = append(existingKeys, k)
		}
	}
	if len(existingKeys) == 0 {
		return
	}
	if err = r.cl.HIncrBy(ctx, sk, spaceSizeKey, -int64(deletedSize)).Err(); err != nil {
		return nil, err
	}

	for _, k := range existingKeys {
		ck := cidKey(k)
		var res int64
		// decrement ref counter for cid
		if res, err = r.cl.Decr(ctx, ck).Result(); err != nil {
			return nil, err
		}
		if res <= 0 {
			if err = r.cl.Del(ctx, ck).Err(); err != nil {
				return nil, err
			}
			toDelete = append(toDelete, cid.MustParse(k))
		}
	}
	return
}

func (r *redisIndex) ExistsInSpace(ctx context.Context, spaceId string, ks []cid.Cid) (exists []cid.Cid, err error) {
	var sk = spaceKey(spaceId)
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

func (r *redisIndex) SpaceSize(ctx context.Context, spaceId string) (size uint64, err error) {
	result, err := r.cl.HGet(ctx, spaceKey(spaceId), spaceSizeKey).Result()
	if err != nil {
		if err == redis.Nil {
			err = nil
		}
		return
	}
	return strconv.ParseUint(result, 10, 64)
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

func spaceKey(spaceId string) string {
	return "s:" + spaceId
}

func cidKey(k string) string {
	return "c:" + k
}

func uniqBlocks(bs []blocks.Block) []blocks.Block {
	if len(bs) <= 1 {
		return bs
	}
	var uniqMap = make(map[string]struct{})
	var res = make([]blocks.Block, 0, len(bs))
	for _, b := range bs {
		if _, ok := uniqMap[b.Cid().KeyString()]; ok {
			continue
		}
		res = append(res, b)
		uniqMap[b.Cid().KeyString()] = struct{}{}
	}
	return res
}

func uniqKeys(ks []cid.Cid) []cid.Cid {
	if len(ks) <= 1 {
		return ks
	}
	var uniqMap = make(map[string]struct{})
	var res = make([]cid.Cid, 0, len(ks))
	for _, k := range ks {
		if _, ok := uniqMap[k.KeyString()]; ok {
			continue
		}
		res = append(res, k)
		uniqMap[k.KeyString()] = struct{}{}
	}
	return res
}
