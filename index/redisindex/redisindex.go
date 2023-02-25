package redisindex

import (
	"context"
	"github.com/anytypeio/any-sync-filenode/index"
	"github.com/anytypeio/any-sync-filenode/redisprovider"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/app/logger"
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
	cl redis.UniversalClient
}

func (r *redisIndex) Init(a *app.App) (err error) {
	r.cl = a.MustComponent(redisprovider.CName).(redisprovider.RedisProvider).Redis()
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
	return res > 0, nil
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
	cidKeys := make([]string, len(bs))
	for i, b := range bs {
		cidKeys[i] = b.Cid().String()
	}

	// add needed cids to space
	toBindKV := make([]interface{}, 0, len(bs)*2)
	for _, b := range bs {
		size := uint64(len(b.RawData()))
		toBindKV = append(toBindKV, b.Cid().String(), Entry{
			Size: size,
			Time: now,
		}.Binary())
	}
	pipe := r.cl.TxPipeline()
	getRes := r.cl.HMGet(ctx, sk, cidKeys...)
	setRes := r.cl.HMSet(ctx, sk, toBindKV...)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	if err := getRes.Err(); err != nil {
		return err
	}
	if err := setRes.Err(); err != nil {
		return err
	}
	getValues := getRes.Val()
	var addedSize uint64
	var toBind = bs[:0]
	for i, b := range bs {
		if getValues[i] == nil {
			addedSize += uint64(len(b.RawData()))
			toBind = append(toBind, b)
		}
	}
	r.cl.HIncrBy(ctx, sk, spaceSizeKey, int64(addedSize))

	// increment ref counter for cids
	for _, b := range toBind {
		if err := r.cl.Incr(ctx, cidKey(b.Cid().String())).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *redisIndex) UnBind(ctx context.Context, spaceId string, ks []cid.Cid) (toDelete []cid.Cid, err error) {
	var sk = spaceKey(spaceId)
	cidKeys := make([]string, len(ks))
	for i, k := range ks {
		cidKeys[i] = k.String()
	}

	pipe := r.cl.TxPipeline()
	// get key
	getRes := pipe.HMGet(ctx, sk, cidKeys...)
	// delete cids from a space
	delRes := pipe.HDel(ctx, sk, cidKeys...)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	if err = getRes.Err(); err != nil {
		return nil, err
	}
	if err = delRes.Err(); err != nil {
		return nil, err
	}
	getVals := getRes.Val()
	var deletedSize uint64
	var existingKeys = cidKeys[:0]
	for i, k := range cidKeys {
		if v := getVals[i]; v != nil {
			if b, ok := v.(string); ok {
				if en, e := NewEntry([]byte(b)); e == nil {
					deletedSize += en.Size
				}
			}
			existingKeys = append(existingKeys, k)
		}
	}

	for i, k := range existingKeys {
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
			toDelete = append(toDelete, ks[i])
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
		return
	}
	return strconv.ParseUint(result, 10, 64)
}

func spaceKey(spaceId string) string {
	return "s:" + spaceId
}

func cidKey(k string) string {
	return "c:" + k
}
