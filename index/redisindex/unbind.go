package redisindex

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type unbindOp struct {
	sk string
	fk string
	ri *redisIndex
}

func (op *unbindOp) Unbind(ctx context.Context) (removedCids []CidInfo, err error) {
	// fetch cids list
	fCids, err := op.ri.newFileCidList(ctx, op.sk, op.fk)
	if err != nil {
		return nil, err
	}
	defer fCids.Unlock()

	// no cids by file, probably it doesn't exist
	if len(fCids.Cids) == 0 {
		return nil, nil
	}

	// fetch cids info
	cids, err := op.ri.cidInfoByByteKeys(ctx, fCids.Cids)
	if err != nil {
		return nil, err
	}
	// additional check
	if len(cids) != len(fCids.Cids) {
		log.Warn("can't fetch all file cids")
	}
	fCids.Cids = nil

	// we need to lock cids here
	cidsToLock := make([]cid.Cid, len(cids))
	for i, c := range cids {
		cidsToLock[i] = c.Cid
	}
	unlock, err := op.ri.Lock(ctx, cidsToLock)
	if err != nil {
		return nil, err
	}
	defer unlock()

	// make a list of results, data will be available after executing of pipeline
	var execResults = make([]*redis.IntCmd, len(cids))
	// do updates in one pipeline
	_, err = op.ri.cl.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, c := range cids {
			execResults[i] = pipe.HIncrBy(ctx, op.sk, c.Cid.String(), -1)
			if err = execResults[i].Err(); err != nil {
				return err
			}
		}
		return fCids.Save(ctx, pipe)
	})

	// check for cids that were removed from a space
	// remove cids without reference
	// make a list with removed cids
	// calculate space size decrease
	var spaceDecreaseSize uint64
	for i, res := range execResults {
		if counter, err := res.Result(); err == nil && counter <= 0 {
			if e := op.ri.cl.HDel(ctx, op.sk, cids[i].Cid.String()).Err(); e != nil {
				log.Warn("can't remove cid from a space", zap.Error(e))
			}
			removedCids = append(removedCids, cids[i])
			spaceDecreaseSize += cids[i].Size_
		}
	}

	// increment space size
	if spaceDecreaseSize > 0 {
		if err = op.ri.cl.HIncrBy(ctx, op.sk, storeSizeKey, -int64(spaceDecreaseSize)).Err(); err != nil {
			return nil, err
		}
	}
	return
}
