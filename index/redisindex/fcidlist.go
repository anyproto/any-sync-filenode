package redisindex

import (
	"bytes"
	"context"
	"github.com/anytypeio/any-sync-filenode/index/redisindex/indexproto"
	"github.com/go-redsync/redsync/v4"
	"github.com/golang/snappy"
	"github.com/ipfs/go-cid"
	"github.com/redis/go-redis/v9"
)

func (r *redisIndex) newFileCidList(ctx context.Context, sk, fk string) (l *fileCidList, err error) {
	lock := r.redsync.NewMutex("_lock:fk:" + sk + ":" + fk)
	if err = lock.LockContext(ctx); err != nil {
		return nil, err
	}
	res, err := r.cl.HGet(ctx, sk, fk).Result()
	if err == redis.Nil {
		err = nil
	}
	if err != nil {
		return
	}
	l = &fileCidList{
		CidList: &indexproto.CidList{},
		cl:      r.cl,
		sk:      sk,
		fk:      fk,
		lock:    lock,
	}
	if len(res) == 0 {
		return
	}
	data, err := snappy.Decode(nil, []byte(res))
	if err != nil {
		return
	}
	if err = l.CidList.Unmarshal(data); err != nil {
		return
	}
	return
}

type fileCidList struct {
	*indexproto.CidList
	cl     redis.UniversalClient
	sk, fk string
	lock   *redsync.Mutex
}

func (cl *fileCidList) Exists(k cid.Cid) bool {
	for _, c := range cl.Cids {
		if bytes.Equal(k.Bytes(), c) {
			return true
		}
	}
	return false
}

func (cl *fileCidList) Add(k cid.Cid) {
	cl.Cids = append(cl.Cids, k.Bytes())
}

func (cl *fileCidList) Unlock() {
	_, _ = cl.lock.Unlock()
}

func (cl *fileCidList) Save(ctx context.Context, tx redis.Pipeliner) (err error) {
	if len(cl.Cids) == 0 {
		return tx.HDel(ctx, cl.sk, cl.fk).Err()
	}
	data, _ := cl.Marshal()
	data = snappy.Encode(nil, data)
	return tx.HSet(ctx, cl.sk, cl.fk, data).Err()
}
