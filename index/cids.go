package index

import (
	"context"
	"errors"
	"time"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/anyproto/any-sync-filenode/index/indexproto"
)

const (
	cidCount      = "cidCount.{system}"
	cidSizeSumKey = "cidSizeSum.{system}"
)

func (ri *redisIndex) CidExists(ctx context.Context, c cid.Cid) (ok bool, err error) {
	return ri.CheckKey(ctx, cidKey(c))
}

func (ri *redisIndex) CidEntries(ctx context.Context, cids []cid.Cid) (entries *CidEntries, err error) {
	entries = &CidEntries{}
	for _, c := range cids {
		if err = ri.getAndAddToEntries(ctx, entries, c); err != nil {
			entries.Release()
			return nil, err
		}
	}
	return entries, nil
}

func (ri *redisIndex) CidEntriesByString(ctx context.Context, cids []string) (entries *CidEntries, err error) {
	entries = &CidEntries{}
	var c cid.Cid
	for _, cs := range cids {
		c, err = cid.Decode(cs)
		if err != nil {
			return
		}
		if err = ri.getAndAddToEntries(ctx, entries, c); err != nil {
			entries.Release()
			return nil, err
		}
	}
	return entries, nil
}

func (ri *redisIndex) CidEntriesByBlocks(ctx context.Context, bs []blocks.Block) (entries *CidEntries, err error) {
	entries = &CidEntries{}
	var visited = make(map[string]struct{})
	for _, b := range bs {
		if _, ok := visited[b.Cid().KeyString()]; ok {
			continue
		}
		if err = ri.getAndAddToEntries(ctx, entries, b.Cid()); err != nil {
			entries.Release()
			return nil, err
		}
		visited[b.Cid().KeyString()] = struct{}{}
	}
	return entries, nil
}

func (ri *redisIndex) getAndAddToEntries(ctx context.Context, entries *CidEntries, c cid.Cid) (err error) {
	_, release, err := ri.AcquireKey(ctx, cidKey(c))
	if err != nil {
		return
	}
	//temporarily ignore the exists check to make a deep check
	/*if !ok {
		release()
		return ErrCidsNotExist
	}*/
	entry, err := ri.getCidEntry(ctx, c)
	if err != nil {
		release()
		return err
	}
	entry.release = release
	entries.entries = append(entries.entries, entry)
	return
}

func (ri *redisIndex) BlocksAdd(ctx context.Context, bs []blocks.Block) (err error) {
	for _, b := range bs {
		exists, release, err := ri.AcquireKey(ctx, cidKey(b.Cid()))
		if err != nil {
			return err
		}
		if !exists {
			if _, err = ri.createCidEntry(ctx, b); err != nil {
				release()
				return err
			}
		} else {
			log.WarnCtx(ctx, "attempt to add existing block", zap.String("cid", b.Cid().String()))
		}
		release()
	}
	return
}

func (ri *redisIndex) CidExistsInSpace(ctx context.Context, k Key, cids []cid.Cid) (exists []cid.Cid, err error) {
	_, release, err := ri.AcquireKey(ctx, spaceKey(k))
	if err != nil {
		return
	}
	defer release()
	var existsRes = make([]*redis.BoolCmd, len(cids))
	_, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, c := range cids {
			existsRes[i] = pipe.HExists(ctx, spaceKey(k), cidKey(c))
		}
		return nil
	})
	for i, c := range cids {
		ex, e := existsRes[i].Result()
		if e != nil {
			return nil, e
		}
		if ex {
			exists = append(exists, c)
		}
	}
	return
}

func (ri *redisIndex) getCidEntry(ctx context.Context, c cid.Cid) (entry *cidEntry, err error) {
	ck := cidKey(c)
	cidData, err := ri.cl.Get(ctx, ck).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// temporary additional check: try to load data from store and restore cid
			var b blocks.Block
			if b, err = ri.persistStore.Get(ctx, c); err != nil {
				log.WarnCtx(ctx, "restore cid entry error", zap.String("cid", c.String()), zap.Error(err))
				err = ErrCidsNotExist
				return
			}
			log.InfoCtx(ctx, "restore cid entry", zap.String("cid", c.String()))
			return ri.createCidEntry(ctx, b)
		}
		return
	}
	protoEntry := &indexproto.CidEntry{}
	err = protoEntry.Unmarshal([]byte(cidData))
	if err != nil {
		return
	}
	entry = &cidEntry{
		Cid:      c,
		CidEntry: protoEntry,
	}
	if err = ri.initCidEntry(ctx, entry); err != nil {
		return nil, err
	}
	return
}

func (ri *redisIndex) createCidEntry(ctx context.Context, b blocks.Block) (entry *cidEntry, err error) {
	now := time.Now().Unix()
	entry = &cidEntry{
		Cid: b.Cid(),
		CidEntry: &indexproto.CidEntry{
			Size_:      uint64(len(b.RawData())),
			CreateTime: now,
			UpdateTime: now,
		},
	}
	if err = ri.initCidEntry(ctx, entry); err != nil {
		return nil, err
	}
	return
}

func (ri *redisIndex) initCidEntry(ctx context.Context, entry *cidEntry) (err error) {
	if entry.Version == 0 {
		entry.Version = 1
		if err != nil {
			return
		}
		if err = entry.Save(ctx, ri.cl); err != nil {
			return
		}
		_, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			if e := pipe.IncrBy(ctx, cidSizeSumKey, int64(entry.Size_)).Err(); e != nil {
				return e
			}
			if e := pipe.Incr(ctx, cidCount).Err(); e != nil {
				return e
			}
			return nil
		})
		return err
	}
	return
}
