package index

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/anyproto/any-sync/commonfile/fileproto/fileprotoerr"
	"github.com/ipfs/go-cid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var ErrLimitExceed = errors.New("limit exceed")

func (ri *redisIndex) CheckLimits(ctx context.Context, key Key) (err error) {
	entry, release, err := ri.AcquireSpace(ctx, key)
	if err != nil {
		return
	}
	defer release()

	// isolated space
	if entry.space.Limit != 0 {
		if entry.space.Size_ >= entry.space.Limit {
			return ErrLimitExceed
		}
		return
	}

	// group limit
	if entry.group.Size_ >= entry.group.Limit {
		return ErrLimitExceed
	}
	return
}

func (ri *redisIndex) SetGroupLimit(ctx context.Context, groupId string, limit uint64) (err error) {
	op := &spaceLimitOp{
		redisIndex: ri,
	}
	return op.SetGroupLimit(ctx, groupId, limit)
}

func (ri *redisIndex) SetSpaceLimit(ctx context.Context, key Key, limit uint64) (err error) {
	op := &spaceLimitOp{
		redisIndex: ri,
	}
	return op.SetSpaceLimit(ctx, key, limit)
}

type spaceLimitOp struct {
	*redisIndex
	groupEntry   *groupEntry
	spaceEntries []*spaceEntry
	release      []func()
}

func (op *spaceLimitOp) SetGroupLimit(ctx context.Context, groupId string, limit uint64) (err error) {
	_, release, err := op.AcquireKey(ctx, groupKey(Key{GroupId: groupId}))
	if err != nil {
		return
	}
	op.release = append(op.release, release)
	defer op.releaseAll()

	op.groupEntry, err = op.getGroupEntry(ctx, Key{GroupId: groupId})
	if err != nil {
		return
	}

	if op.groupEntry.AccountLimit == limit {
		// same limit - do nothing
		return
	}

	isolatedLimit := op.groupEntry.AccountLimit - op.groupEntry.Limit
	if op.groupEntry.AccountLimit > limit {
		// limit is decreasing
		// check the isolated limit, if the new limit is covering the isolated limit - do nothing
		if isolatedLimit > limit {
			// decrease the limit for isolated spaces in the same proportion as the account limit
			isolatedLimit, err = op.decreaseIsolatedLimit(ctx, float64(limit)/float64(op.groupEntry.AccountLimit))
			if err != nil {
				return err
			}
		}
	}

	op.groupEntry.Limit = limit - isolatedLimit
	op.groupEntry.AccountLimit = limit
	return op.saveAll(ctx)
}

func (op *spaceLimitOp) decreaseIsolatedLimit(ctx context.Context, k float64) (newIsolatedLimit uint64, err error) {
	for _, spaceId := range op.groupEntry.SpaceIds {
		newLimit, err := op.decreaseIsolatedLimitForSpace(ctx, spaceId, k)
		if err != nil {
			return 0, err
		}
		newIsolatedLimit += newLimit
	}
	return
}

func (op *spaceLimitOp) decreaseIsolatedLimitForSpace(ctx context.Context, spaceId string, k float64) (newIsolatedLimit uint64, err error) {
	key := Key{GroupId: op.groupEntry.GroupId, SpaceId: spaceId}
	_, release, err := op.AcquireKey(ctx, spaceKey(key))
	if err != nil {
		return
	}
	op.release = append(op.release, release)

	sEntry, err := op.getSpaceEntry(ctx, key)
	if err != nil {
		return
	}
	if sEntry.Limit == 0 {
		// not isolated space
		return
	}

	sEntry.Limit = uint64(float64(sEntry.Limit) * k)
	op.spaceEntries = append(op.spaceEntries, sEntry)
	return sEntry.Limit, nil
}

func (op *spaceLimitOp) SetSpaceLimit(ctx context.Context, key Key, limit uint64) (err error) {
	entry, release, err := op.AcquireSpace(ctx, key)
	if err != nil {
		return
	}
	op.release = append(op.release, release)
	defer op.releaseAll()

	op.groupEntry = entry.group
	op.spaceEntries = append(op.spaceEntries, entry.space)

	if entry.space.Limit == limit {
		// same limit - do nothing
		return
	}
	prevLimit := entry.space.Limit
	entry.space.Limit = limit

	if limit != 0 {
		diff := int64(limit) - int64(prevLimit)
		newGLimit := int64(op.groupEntry.Limit) - diff

		// check the group limit
		if newGLimit < int64(op.groupEntry.Size_) {
			return fileprotoerr.ErrNotEnoughSpace
		}

		// check the space limit
		if entry.space.Size_ > limit {
			return fileprotoerr.ErrNotEnoughSpace
		}
		op.groupEntry.Limit = uint64(newGLimit)

		if prevLimit == 0 {
			if err = op.isolateSpace(ctx, entry); err != nil {
				return
			}
		}
	} else {
		if err = op.uniteSpace(ctx, entry); err != nil {
			return
		}
		op.groupEntry.Limit += prevLimit
	}
	entry.group.AddSpaceId(key.SpaceId)
	return op.saveAll(ctx)
}

func (op *spaceLimitOp) isolateSpace(ctx context.Context, entry groupSpaceEntry) (err error) {
	key := Key{GroupId: entry.group.GroupId, SpaceId: entry.space.Id}
	sk := spaceKey(key)
	gk := groupKey(key)

	keys, err := op.cl.HGetAll(ctx, sk).Result()
	if err != nil {
		return
	}

	var cids []cid.Cid
	var cidRefs []int64
	// collect all cids
	for k, val := range keys {
		if strings.HasPrefix(k, "c:") {
			c, cErr := cid.Decode(k[2:])
			if cErr != nil {
				log.WarnCtx(ctx, "can't decode cid", zap.String("cid", k[2:]), zap.Error(cErr))
			} else {
				ref, _ := strconv.ParseInt(val, 10, 64)
				cids = append(cids, c)
				cidRefs = append(cidRefs, ref)
			}
		}
	}

	// fetch cid entries
	// TODO: we don't need to take a lock here, but for now it easiest way
	cidEntries, err := op.CidEntries(ctx, cids)
	if err != nil {
		return err
	}
	defer cidEntries.Release()

	// decrement refs in the group
	var groupDecrResults = make([]*redis.IntCmd, len(cids))
	if _, err = op.cl.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, c := range cids {
			groupDecrResults[i] = pipe.HIncrBy(ctx, gk, cidKey(c), -cidRefs[i])
		}
		return nil
	}); err != nil {
		return
	}

	// make a list of cids without refs
	var toDeleteIdx = make([]int, 0, len(cids))
	for i, res := range groupDecrResults {
		if res.Val() <= 0 {
			toDeleteIdx = append(toDeleteIdx, i)
		}
	}

	// remove cids without refs from the group and decrease the size
	if len(toDeleteIdx) > 0 {
		if _, err = op.cl.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			for _, idx := range toDeleteIdx {
				pipe.HDel(ctx, gk, cidKey(cids[idx]))
				op.groupEntry.Size_ -= cidEntries.entries[idx].Size_
				op.groupEntry.CidCount--
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return
}

func (op *spaceLimitOp) uniteSpace(ctx context.Context, entry groupSpaceEntry) (err error) {
	key := Key{GroupId: entry.group.GroupId, SpaceId: entry.space.Id}
	sk := spaceKey(key)
	gk := groupKey(key)

	keys, err := op.cl.HKeys(ctx, sk).Result()
	if err != nil {
		return
	}

	var cids []cid.Cid
	// collect all cids
	for _, k := range keys {
		if strings.HasPrefix(k, "c:") {
			c, cErr := cid.Decode(k[2:])
			if cErr != nil {
				log.WarnCtx(ctx, "can't decode cid", zap.String("cid", k[2:]), zap.Error(cErr))
			} else {
				cids = append(cids, c)
			}
		}
	}

	// fetch cid entries
	// TODO: we don't need to take a lock here, but for now it easiest way
	cidEntries, err := op.CidEntries(ctx, cids)
	if err != nil {
		return err
	}
	defer cidEntries.Release()

	// increment refs in the group
	var groupDecrResults = make([]*redis.IntCmd, len(cids))
	if _, err = op.cl.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, c := range cids {
			groupDecrResults[i] = pipe.HIncrBy(ctx, gk, cidKey(c), 1)
		}
		return nil
	}); err != nil {
		return
	}

	// make a list of cids without refs and increase the size
	for i, res := range groupDecrResults {
		if res.Val() == 1 {
			op.groupEntry.Size_ += cidEntries.entries[i].Size_
			op.groupEntry.CidCount++
		}
	}
	return
}

func (op *spaceLimitOp) releaseAll() {
	for _, r := range op.release {
		r()
	}
}

func (op *spaceLimitOp) saveAll(ctx context.Context) (err error) {
	_, err = op.cl.TxPipelined(ctx, func(tx redis.Pipeliner) error {
		for _, sEntry := range op.spaceEntries {
			sEntry.Save(ctx, Key{GroupId: sEntry.GroupId, SpaceId: sEntry.Id}, tx)
		}
		op.groupEntry.Save(ctx, tx)
		return nil
	})
	return
}
