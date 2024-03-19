package index

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/go-redsync/redsync/v4"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	partitionCount = 256

	persistThreads = 10
)

var partitions = make([]int, partitionCount)

func init() {
	for i := range partitions {
		partitions[i] = i
	}
}

type persistentStore interface {
	IndexGet(ctx context.Context, key string) (value []byte, err error)
	IndexPut(ctx context.Context, key string, value []byte) (err error)

	Get(ctx context.Context, k cid.Cid) (blocks.Block, error)
}

func bloomFilterKey(key string) string {
	sum := xxhash.Sum64String(key) % partitionCount
	return "bf:{" + strconv.FormatUint(sum, 10) + "}"
}

func storeKey(key string) string {
	sum := xxhash.Sum64String(key) % partitionCount
	return "store:{" + strconv.FormatUint(sum, 10) + "}"
}

func (ri *redisIndex) CheckKey(ctx context.Context, key string) (exists bool, err error) {
	var release func()
	if exists, release, err = ri.acquireKey(ctx, key); err != nil {
		return
	}
	if exists {
		if err = ri.updateKeyUsage(ctx, key); err != nil {
			release()
			return false, err
		}
	}
	release()
	return
}

func (ri *redisIndex) AcquireKey(ctx context.Context, key string) (exists bool, release func(), err error) {
	if exists, release, err = ri.acquireKey(ctx, key); err != nil {
		return
	}
	if err = ri.updateKeyUsage(ctx, key); err != nil {
		release()
		return false, nil, err
	}
	return
}

func (ri *redisIndex) AcquireSpace(ctx context.Context, key Key) (entry groupSpaceEntry, release func(), err error) {
	gExists, gRelease, err := ri.AcquireKey(ctx, groupKey(key))
	if err != nil {
		return
	}
	sExists, sRelease, err := ri.AcquireKey(ctx, spaceKey(key))
	if err != nil {
		gRelease()
		return
	}

	if entry.space, err = ri.getSpaceEntry(ctx, key); err != nil {
		gRelease()
		sRelease()
		return
	}
	if entry.group, err = ri.getGroupEntry(ctx, key); err != nil {
		gRelease()
		sRelease()
		return
	}

	if entry.space.GetGroupId() != key.GroupId {
		gRelease()
		sRelease()
		err = fmt.Errorf("space and group mismatched")
		return
	}

	release = func() {
		gRelease()
		sRelease()
	}
	entry.spaceExists = sExists
	entry.groupExists = gExists

	return
}

func (ri *redisIndex) acquireKey(ctx context.Context, key string) (exists bool, release func(), err error) {
	mu := ri.redsync.NewMutex("_lock:"+key, redsync.WithExpiry(time.Minute*20))
	if err = mu.LockContext(ctx); err != nil {
		return
	}
	release = func() {
		_, _ = mu.Unlock()
	}

	// check in redis
	ex, err := ri.cl.Exists(ctx, key).Result()
	if err != nil {
		release()
		return
	}
	// already in redis
	if ex > 0 {
		return true, release, nil
	}

	// check bloom filter
	bfKey := bloomFilterKey(key)
	bloomEx, err := ri.cl.BFExists(ctx, bfKey, key).Result()
	if err != nil {
		release()
		return false, nil, err
	}
	// not in bloom filter, item not exists
	if !bloomEx {
		return false, release, nil
	}

	// try to load from persistent store
	val, err := ri.persistStore.IndexGet(ctx, key)
	if err != nil {
		release()
		return false, nil, err
	}
	// nil means not found
	if val == nil {
		return false, release, nil
	}
	if err = ri.cl.Restore(ctx, key, 0, string(val)).Err(); err != nil {
		release()
		return false, nil, err
	}
	return true, release, nil
}
func (ri *redisIndex) updateKeyUsage(ctx context.Context, key string) (err error) {
	sKey := storeKey(key)
	return ri.cl.ZAdd(ctx, sKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: key,
	}).Err()
}

func (ri *redisIndex) PersistKeys(ctx context.Context) {
	st := time.Now()
	rand.Shuffle(len(partitions), func(i, j int) {
		partitions[i], partitions[j] = partitions[j], partitions[i]
	})
	stat := &persistStat{}
	var wg sync.WaitGroup
	wg.Add(len(partitions))
	var limiter = make(chan struct{}, persistThreads)
	for _, part := range partitions {
		limiter <- struct{}{}
		go func(p int) {
			defer func() {
				<-limiter
				wg.Done()
			}()
			if e := ri.persistKeys(ctx, p, stat); e != nil {
				log.Warn("persist part error", zap.Error(e), zap.Int("part", p))
			}
		}(part)
	}
	wg.Wait()
	log.Info("persist",
		zap.Duration("dur", time.Since(st)),
		zap.Int32("handled", stat.handled.Load()),
		zap.Int32("deleted", stat.deleted.Load()),
		zap.Int32("missed", stat.missed.Load()),
		zap.Int32("errors", stat.errors.Load()),
		zap.Int32("moved", stat.moved.Load()),
		zap.Int32("moved kbs", stat.movedBytes.Load()/1024),
	)
}

func (ri *redisIndex) persistKeys(ctx context.Context, part int, stat *persistStat) (err error) {
	deadline := time.Now().Add(-ri.persistTtl).Unix()
	sk := "store:{" + strconv.FormatInt(int64(part), 10) + "}"
	keys, err := ri.cl.ZRangeByScore(ctx, sk, &redis.ZRangeBy{
		Min: "0",
		Max: strconv.FormatInt(deadline, 10),
	}).Result()
	if err != nil {
		return
	}
	for _, k := range keys {
		if err = ri.persistKey(ctx, sk, k, deadline, stat); err != nil {
			return
		}
	}
	return
}

func (ri *redisIndex) persistKey(ctx context.Context, storeKey, key string, deadline int64, stat *persistStat) (err error) {
	mu := ri.redsync.NewMutex("_lock:" + key)
	if err = mu.LockContext(ctx); err != nil {
		return
	}
	defer func() {
		_, _ = mu.Unlock()
	}()

	stat.handled.Add(1)

	// make sure lastActivity not changed
	res, err := ri.cl.ZMScore(ctx, storeKey, key).Result()
	if err != nil {
		stat.errors.Add(1)
		return
	}
	if int64(res[0]) > deadline || res[0] == 0 {
		stat.missed.Add(1)
		return
	}

	dump, err := ri.cl.Dump(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		stat.errors.Add(1)
		return
	}
	// key was removed - just remove it from store queue
	if errors.Is(err, redis.Nil) {
		stat.deleted.Add(1)
		return ri.cl.ZRem(ctx, storeKey, key).Err()
	}

	// persist the dump
	if err = ri.persistStore.IndexPut(ctx, key, []byte(dump)); err != nil {
		return
	}
	// remove from queue and add to bloom filter
	_, err = ri.cl.TxPipelined(ctx, func(tx redis.Pipeliner) error {
		tx.ZRem(ctx, storeKey, key)
		tx.BFAdd(ctx, bloomFilterKey(key), key)
		return nil
	})

	stat.moved.Add(1)
	stat.movedBytes.Add(int32(len(dump)))

	// remove key
	return ri.cl.Del(ctx, key).Err()
}

type persistStat struct {
	handled    atomic.Int32
	moved      atomic.Int32
	movedBytes atomic.Int32
	missed     atomic.Int32
	deleted    atomic.Int32
	errors     atomic.Int32
}
