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
	"time"
)

const CName = "filenode.redisindex"

var log = logger.NewNamed(CName)

func New() RedisIndex {
	return new(redisIndex)
}

type RedisIndex interface {
	index.Index
	app.Component
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
	// check cids existence in space
	result, err := r.cl.HMGet(ctx, sk, cidKeys...).Result()
	if err != nil {
		return err
	}
	toBind := bs[:0]
	for i, b := range bs {
		if result[i] == nil {
			toBind = append(toBind, b)
		}
	}

	// all cids exists in a space - nothing to do
	if len(toBind) == 0 {
		return nil
	}

	// add needed cids to space
	toBindKV := make([]interface{}, 0, len(toBind)*2)
	for _, b := range toBind {
		toBindKV = append(toBindKV, b.Cid().String(), Entry{
			Size: uint64(len(b.RawData())),
			Time: now,
		}.Binary())
	}
	if err = r.cl.HMSet(ctx, sk, toBindKV...).Err(); err != nil {
		return err
	}

	// add space to cids
	for _, b := range toBind {
		if err = r.cl.LPush(ctx, cidKey(b.Cid().String()), spaceId).Err(); err != nil {
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
	// delete cids from a space
	deleted, err := r.cl.HDel(ctx, sk, cidKeys...).Result()
	if err != nil {
		return nil, err
	}
	if deleted == 0 {
		// nothing to delete
		return
	}
	var checkDeleteKeys = make([]string, 0, len(cidKeys))
	for _, k := range cidKeys {
		ck := cidKey(k)
		var res int64
		// remove space from cids
		if res, err = r.cl.LRem(ctx, ck, 0, spaceId).Result(); err != nil {
			return nil, err
		}
		if res > 0 {
			// if we successfully removed space, mark key to check for empty
			checkDeleteKeys = append(checkDeleteKeys, ck)
		}
	}

	for _, k := range checkDeleteKeys {
		var l int64
		if l, err = r.cl.LLen(ctx, k).Result(); err != nil {
			return nil, err
		}
		if l == 0 {
			toDelete = append(toDelete, cid.MustParse(k[2:]))
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
		return nil, err
	}
	exists = make([]cid.Cid, 0, len(ks))
	for i, v := range result {
		if v != nil {
			exists = append(exists, ks[i])
		}
	}
	return
}

func spaceKey(spaceId string) string {
	return "s:" + spaceId
}

func cidKey(k string) string {
	return "c:" + k
}
