package index

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func (ri *redisIndex) FileUnbind(ctx context.Context, key Key, fileIds ...string) (removedCids []cid.Cid, err error) {
	entry, release, err := ri.AcquireSpace(ctx, key)
	if err != nil {
		return
	}
	defer release()
	for _, fileId := range fileIds {
		rem, err := ri.fileUnbind(ctx, key, entry, fileId)
		if err != nil {
			return removedCids, err
		}
		removedCids = append(removedCids, rem...)
	}
	return
}

func (ri *redisIndex) fileUnbind(ctx context.Context, key Key, entry groupSpaceEntry, fileId string) (removedCids []cid.Cid, err error) {
	var (
		sk = SpaceKey(key)
		gk = GroupKey(key)
	)
	// get file entry
	fileInfo, isNewFile, err := ri.getFileEntry(ctx, key, fileId)
	if err != nil {
		return
	}
	if isNewFile {
		// means file doesn't exist
		return nil, nil
	}

	// fetch cids
	cids, err := ri.CidEntriesByString(ctx, fileInfo.Cids)
	if err != nil {
		return nil, err
	}
	defer cids.Release()

	isolatedSpace := entry.space.Limit != 0

	// fetch cid refs in one pipeline
	var (
		groupCidRefs = make([]*redis.StringCmd, len(cids.entries))
		spaceCidRefs = make([]*redis.StringCmd, len(cids.entries))
	)
	_, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, c := range cids.entries {
			if !isolatedSpace {
				groupCidRefs[i] = pipe.HGet(ctx, gk, CidKey(c.Cid))
			}
			spaceCidRefs[i] = pipe.HGet(ctx, sk, CidKey(c.Cid))
		}
		return nil
	})
	if err != nil {
		return
	}

	// update info and calculate changes
	var (
		groupRemoveKeys = make([]string, 0, len(cids.entries))
		spaceRemoveKeys = make([]string, 0, len(cids.entries))
		groupDecrKeys   = make([]string, 0, len(cids.entries))
		spaceDecrKeys   = make([]string, 0, len(cids.entries))
		affectedCidIdx  = make([]int, 0, len(cids.entries))
	)
	if entry.space.FileCount != 0 {
		entry.space.FileCount--
	} else {
		log.WarnCtx(ctx, "file: unable to decrement 0-ref", zap.String("spaceId", key.SpaceId))
	}
	for i, c := range cids.entries {
		ck := CidKey(c.Cid)
		if !isolatedSpace {
			res, err := groupCidRefs[i].Result()
			if err != nil {
				return nil, err
			}
			if res == "1" {
				groupRemoveKeys = append(groupRemoveKeys, ck)
				if entry.group.Size_-c.Size_ > entry.group.Size_ {
					log.WarnCtx(ctx, "group: unable to decrement size", zap.Uint64("before", entry.group.Size_), zap.Uint64("size", c.Size_), zap.String("spaceId", key.SpaceId))
				} else {
					entry.group.Size_ -= c.Size_
				}
				if entry.group.CidCount != 0 {
					entry.group.CidCount--
				} else {
					log.WarnCtx(ctx, "group: unable to decrement 0-ref", zap.String("spaceId", key.SpaceId))
				}
			} else {
				groupDecrKeys = append(groupDecrKeys, ck)
			}
		}
		res, err := spaceCidRefs[i].Result()
		if err != nil {
			return nil, err
		}
		if res == "1" {
			spaceRemoveKeys = append(spaceRemoveKeys, ck)
			if entry.space.Size_-c.Size_ > entry.space.Size_ {
				log.WarnCtx(ctx, "space: unable to decrement size", zap.Uint64("before", entry.space.Size_), zap.Uint64("size", c.Size_), zap.String("spaceId", key.SpaceId))
			} else {
				entry.space.Size_ -= c.Size_
			}
			if entry.space.CidCount != 0 {
				entry.space.CidCount--
			} else {
				log.WarnCtx(ctx, "space: unable to decrement 0-ref", zap.String("spaceId", key.SpaceId))
			}
			affectedCidIdx = append(affectedCidIdx, i)
		} else {
			spaceDecrKeys = append(spaceDecrKeys, ck)
		}
	}

	// do updates in one tx
	_, err = ri.cl.TxPipelined(ctx, func(tx redis.Pipeliner) error {
		tx.HDel(ctx, sk, FileKey(fileId))
		if len(spaceRemoveKeys) != 0 {
			tx.HDel(ctx, sk, spaceRemoveKeys...)
		}
		if len(groupRemoveKeys) != 0 {
			tx.HDel(ctx, gk, groupRemoveKeys...)
		}
		if len(spaceDecrKeys) != 0 {
			for _, k := range spaceDecrKeys {
				tx.HIncrBy(ctx, sk, k, -1)
			}
		}
		if len(groupDecrKeys) != 0 {
			for _, k := range groupDecrKeys {
				tx.HIncrBy(ctx, gk, k, -1)
			}
		}
		entry.space.Save(ctx, key, tx)
		entry.group.Save(ctx, tx)
		return nil
	})

	// update cids
	for _, idx := range affectedCidIdx {
		if cids.entries[idx].Refs != 0 {
			cids.entries[idx].Refs--
		} else {
			log.WarnCtx(ctx, "cid: unable to decrement 0-ref", zap.String("cid", cids.entries[idx].Cid.String()), zap.String("spaceId", key.SpaceId))
			continue
		}
		if cids.entries[idx].Refs == 0 {
			removedCids = append(removedCids, cids.entries[idx].Cid)
		}
		if saveErr := cids.entries[idx].Save(ctx, ri.cl); saveErr != nil {
			log.WarnCtx(ctx, "unable to save cid info", zap.Error(saveErr), zap.String("cid", cids.entries[idx].Cid.String()), zap.String("spaceId", key.SpaceId))
		}
	}
	return
}
