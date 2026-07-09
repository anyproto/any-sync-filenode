package index

import (
	"context"
	"errors"
	"time"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/anyproto/any-sync-filenode/index/indexproto"
)

const (
	cidCount      = "cidCount.{system}"
	cidSizeSumKey = "cidSizeSum.{system}"
)

func (ri *redisIndex) CidExists(ctx context.Context, c cid.Cid) (ok bool, err error) {
	return ri.CheckKey(ctx, CidKey(c))
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

func (ri *redisIndex) CidEntriesByString(ctx context.Context, cids []string) (*CidEntries, error) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	results := make([]*cidEntry, len(cids))

	for i, cs := range cids {
		i, cs := i, cs
		g.Go(func() error {
			// decode
			c, err := cid.Decode(cs)
			if err != nil {
				return err
			}
			// fetch entry
			entry, err := ri.acquireCidEntry(ctx, c)
			if err != nil {
				return err
			}
			// store in the correct slot
			results[i] = entry
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		for _, result := range results {
			if result != nil {
				result.release()
			}
		}
		return nil, err
	}

	entries := &CidEntries{}
	for _, e := range results {
		entries.Add(e)
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
	entry, err := ri.acquireCidEntry(ctx, c)
	if err != nil {
		return err
	}
	entries.Add(entry)
	return
}

func (ri *redisIndex) acquireCidEntry(ctx context.Context, c cid.Cid) (entry *cidEntry, err error) {
	_, release, err := ri.AcquireKey(ctx, CidKey(c))
	if err != nil {
		return
	}

	entry, err = ri.getCidEntry(ctx, c)
	if err != nil {
		release()
		return nil, err
	}
	entry.release = release
	return
}

func (ri *redisIndex) BlocksAdd(ctx context.Context, bs []blocks.Block) (err error) {
	for _, b := range bs {
		exists, release, err := ri.AcquireKey(ctx, CidKey(b.Cid()))
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

// CidExistsBulk returns, for each input CID, whether it exists in the global
// index (either live in Redis or in persistent store). It is optimized for
// BlocksCheck: no distributed locks, no usage-tracker writes, no restore of
// evicted keys. Order of the returned slice matches the input order.
func (ri *redisIndex) CidExistsBulk(ctx context.Context, cids []cid.Cid) (exists []bool, err error) {
	if len(cids) == 0 {
		return nil, nil
	}
	exists = make([]bool, len(cids))

	// Phase A: pipelined EXISTS against per-CID keys in Redis.
	existsCmds := make([]*redis.IntCmd, len(cids))
	if _, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, c := range cids {
			existsCmds[i] = pipe.Exists(ctx, CidKey(c))
		}
		return nil
	}); err != nil {
		return nil, err
	}

	type missed struct {
		idx int
		key string
	}
	missedByPart := make(map[string][]missed)
	for i, c := range cids {
		n, e := existsCmds[i].Result()
		if e != nil {
			return nil, e
		}
		if n > 0 {
			exists[i] = true
			continue
		}
		key := CidKey(c)
		bfKey := bloomFilterKey(key)
		missedByPart[bfKey] = append(missedByPart[bfKey], missed{idx: i, key: key})
	}
	if len(missedByPart) == 0 {
		return exists, nil
	}

	// Phase B: pipelined BFMEXISTS per partition.
	bfCmds := make(map[string]*redis.BoolSliceCmd, len(missedByPart))
	if _, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for bfKey, ms := range missedByPart {
			items := make([]interface{}, len(ms))
			for i, m := range ms {
				items[i] = m.key
			}
			bfCmds[bfKey] = pipe.BFMExists(ctx, bfKey, items...)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	type candidate struct {
		idx int
		key string
	}
	var candidates []candidate
	for bfKey, ms := range missedByPart {
		bfRes, e := bfCmds[bfKey].Result()
		if e != nil {
			return nil, e
		}
		for i, hit := range bfRes {
			if hit {
				candidates = append(candidates, candidate{idx: ms[i].idx, key: ms[i].key})
			}
		}
	}
	if len(candidates) == 0 {
		return exists, nil
	}

	// Phase C: parallel persistStore fallback for bloom-positive candidates.
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(10)
	for _, cand := range candidates {
		cand := cand
		g.Go(func() error {
			val, e := ri.persistStore.IndexGet(gctx, cand.key)
			if e != nil {
				return e
			}
			if val != nil {
				exists[cand.idx] = true
			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		return nil, err
	}
	return exists, nil
}

func (ri *redisIndex) CidExistsInSpace(ctx context.Context, k Key, cids []cid.Cid) (exists []cid.Cid, err error) {
	_, release, err := ri.AcquireKey(ctx, SpaceKey(k))
	if err != nil {
		return
	}
	defer release()
	var existsRes = make([]*redis.BoolCmd, len(cids))
	_, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, c := range cids {
			existsRes[i] = pipe.HExists(ctx, SpaceKey(k), CidKey(c))
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
	ck := CidKey(c)
	cidData, err := ri.cl.Get(ctx, ck).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
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
	err = protoEntry.UnmarshalVT([]byte(cidData))
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
			Size:       uint64(len(b.RawData())),
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
			if e := pipe.IncrBy(ctx, cidSizeSumKey, int64(entry.Size)).Err(); e != nil {
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
