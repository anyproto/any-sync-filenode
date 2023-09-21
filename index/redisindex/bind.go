package redisindex

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type bindOp struct {
	sk string
	fk string
	ri *redisIndex
}

func (op *bindOp) Bind(ctx context.Context, cids []CidInfo) (newCids []CidInfo, err error) {
	// fetch existing or create new file cids list
	fCids, err := op.ri.newFileCidList(ctx, op.sk, op.fk)
	if err != nil {
		return nil, err
	}
	defer fCids.Unlock()

	// make a list of indexes non-exists cids
	var newInFileIdx = make([]int, 0, len(cids))
	for i, c := range cids {
		if !fCids.Exists(c.Cid) {
			newInFileIdx = append(newInFileIdx, i)
			fCids.Add(c.Cid)
		}
	}

	// all cids exists, nothing to do
	if len(newInFileIdx) == 0 {
		return nil, nil
	}

	// make a list of results, data will be available after executing of pipeline
	var execResults = make([]*redis.IntCmd, len(newInFileIdx))
	// do updates in one pipeline
	_, err = op.ri.cl.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, idx := range newInFileIdx {
			ck := cids[idx].Cid.String()
			execResults[i] = pipe.HIncrBy(ctx, op.sk, ck, 1)
			if err = execResults[i].Err(); err != nil {
				return err
			}
		}
		return fCids.Save(ctx, pipe)
	})
	if err != nil {
		return
	}

	// check for newly added cids
	// calculate spaceSize increase
	// make a list of finally added cids
	var spaceIncreaseSize uint64
	for i, res := range execResults {
		if newCounter, err := res.Result(); err == nil && newCounter == 1 {
			c := cids[newInFileIdx[i]]
			spaceIncreaseSize += c.Size_
			newCids = append(newCids, c)
		}
	}

	// increment space size
	if spaceIncreaseSize > 0 {
		if err = op.ri.cl.HIncrBy(ctx, op.sk, storeSizeKey, int64(spaceIncreaseSize)).Err(); err != nil {
			return nil, err
		}
	}
	return
}
