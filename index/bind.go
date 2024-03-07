package index

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func (ri *redisIndex) FileBind(ctx context.Context, key Key, fileId string, cids *CidEntries) (err error) {
	entry, release, err := ri.AcquireSpace(ctx, key)
	if err != nil {
		return
	}
	defer release()

	return ri.fileBind(ctx, key, fileId, cids, entry)
}

func (ri *redisIndex) fileBind(ctx context.Context, key Key, fileId string, cids *CidEntries, entry groupSpaceEntry) (err error) {
	var gk = groupKey(key)
	var sk = spaceKey(key)

	// get file entry
	fileInfo, isNewFile, err := ri.getFileEntry(ctx, key, fileId)
	if err != nil {
		return
	}

	isolatedSpace := entry.space.Limit != 0

	// make a list of indexes of non-exists cids
	var newFileCidIdx = make([]int, 0, len(cids.entries))
	for i, c := range cids.entries {
		if !fileInfo.Exists(c.Cid.String()) {
			newFileCidIdx = append(newFileCidIdx, i)
			fileInfo.Cids = append(fileInfo.Cids, c.Cid.String())
			fileInfo.Size_ += c.Size_
		}
	}

	// all cids exists, nothing to do
	if len(newFileCidIdx) == 0 {
		return
	}

	var affectedCidIdx = make([]int, 0, len(cids.entries))

	// get all cids from space and group in one pipeline
	var (
		cidExistSpaceCmds = make([]*redis.BoolCmd, len(newFileCidIdx))
		cidExistGroupCmds = make([]*redis.BoolCmd, len(newFileCidIdx))
	)
	_, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, idx := range newFileCidIdx {
			ck := cidKey(cids.entries[idx].Cid)
			cidExistSpaceCmds[i] = pipe.HExists(ctx, sk, ck)
			if !isolatedSpace {
				cidExistGroupCmds[i] = pipe.HExists(ctx, gk, ck)
			}
		}
		return nil
	})
	if err != nil {
		return
	}

	// calculate new group and space stats
	for i, idx := range newFileCidIdx {
		if !isolatedSpace {
			ex, err := cidExistGroupCmds[i].Result()
			if err != nil {
				return err
			}
			if !ex {
				entry.group.CidCount++
				entry.group.Size_ += cids.entries[idx].Size_
			}
		}
		ex, err := cidExistSpaceCmds[i].Result()
		if err != nil {
			return err
		}
		if !ex {
			entry.space.CidCount++
			entry.space.Size_ += cids.entries[idx].Size_
			affectedCidIdx = append(affectedCidIdx, idx)
		}
	}
	entry.group.AddSpaceId(key.SpaceId)
	if isNewFile {
		entry.space.FileCount++
	}

	// make group and space updates in one tx
	_, err = ri.cl.TxPipelined(ctx, func(tx redis.Pipeliner) error {
		// increment cid refs
		for _, idx := range newFileCidIdx {
			ck := cidKey(cids.entries[idx].Cid)
			if !isolatedSpace {
				tx.HIncrBy(ctx, gk, ck, 1)
			}
			tx.HIncrBy(ctx, sk, ck, 1)
		}
		// save info
		entry.space.Save(ctx, key, tx)
		entry.group.Save(ctx, tx)
		fileInfo.Save(ctx, key, fileId, tx)
		return nil
	})

	// update cids
	for _, idx := range affectedCidIdx {
		cids.entries[idx].Refs++
		if saveErr := cids.entries[idx].Save(ctx, ri.cl); saveErr != nil {
			log.WarnCtx(ctx, "unable to save cid info", zap.Error(saveErr), zap.String("cid", cids.entries[idx].Cid.String()))
		}
	}
	return
}
