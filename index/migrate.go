package index

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/golang/snappy"
	"github.com/ipfs/go-cid"
	"go.uber.org/zap"

	"github.com/anyproto/any-sync-filenode/index/indexproto"
)

// Migrate from the version without groups
func (ri *redisIndex) Migrate(ctx context.Context, key Key) (err error) {
	st := time.Now()
	// fast check before lock the key
	var checkKeys = []string{
		"s:" + key.SpaceId,
		"s:" + key.GroupId,
	}
	var migrateKey string
	for _, checkKey := range checkKeys {
		ex, err := ri.cl.Exists(ctx, checkKey).Result()
		if err != nil {
			return err
		}
		if ex != 0 {
			migrateKey = checkKey
			break
		}
	}
	// old keys doesn't exist
	if migrateKey == "" {
		return
	}

	// lock the key
	mu := ri.redsync.NewMutex("_lock:"+migrateKey, redsync.WithExpiry(time.Minute))
	if err = mu.LockContext(ctx); err != nil {
		return
	}
	defer func() {
		_, _ = mu.Unlock()
	}()

	// another check under lock
	ex, err := ri.cl.Exists(ctx, migrateKey).Result()
	if err != nil {
		return
	}
	if ex == 0 {
		// key doesn't exist - another node did the migration
		return
	}

	keys, err := ri.cl.HKeys(ctx, migrateKey).Result()
	if err != nil {
		return
	}

	for _, k := range keys {
		if strings.HasPrefix(k, "f:") {
			if err = ri.migrateFile(ctx, key, migrateKey, k); err != nil {
				if errors.Is(err, ErrCidsNotExist) {
					err = nil
					log.WarnCtx(ctx, "migrate cid no exists", zap.String("spaceId", key.SpaceId), zap.String("fileId", k))
					continue
				}
				return
			}
		}
	}
	if err = ri.cl.Del(ctx, migrateKey).Err(); err != nil {
		return
	}
	log.Info("space migrated", zap.String("spaceId", key.SpaceId), zap.Duration("dur", time.Since(st)))
	return
}

func (ri *redisIndex) migrateFile(ctx context.Context, key Key, migrateKey, fileKey string) (err error) {
	data, err := ri.cl.HGet(ctx, migrateKey, fileKey).Result()
	if err != nil {
		return
	}

	encodedData, err := snappy.Decode(nil, []byte(data))
	if err != nil {
		return
	}
	oldList := &indexproto.CidList{}
	if err = oldList.Unmarshal(encodedData); err != nil {
		return
	}

	var cidList = make([]cid.Cid, len(oldList.Cids))
	for i, cb := range oldList.Cids {
		if cidList[i], err = cid.Cast(cb); err != nil {
			return
		}
	}

	cidEntries, err := ri.CidEntries(ctx, cidList)
	if err != nil {
		return
	}
	defer cidEntries.Release()

	return ri.FileBind(ctx, key, fileKey[2:], cidEntries)
}
