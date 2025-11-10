package index

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/redis/go-redis/v9"
)

/**
TODO: make a separate keys with status and history of ownership
*/

func (ri *redisIndex) Move(ctx context.Context, dest, src Key) (err error) {
	if dest.SpaceId != src.SpaceId {
		return fmt.Errorf("spaceIs should be the same for both keys")
	}
	srcEntry, scrRelease, err := ri.AcquireSpace(ctx, src)
	if err != nil {
		return
	}
	defer scrRelease()

	_, destGRelease, err := ri.AcquireKey(ctx, GroupKey(dest))
	if err != nil {
		return
	}
	defer destGRelease()

	destGroup, err := ri.getGroupEntry(ctx, dest)
	if err != nil {
		return
	}

	// in case of hash collision, we don't need to lock the space key twice
	sSK := SpaceKey(src)
	dSK := SpaceKey(dest)
	if sSK != dSK {
		_, destSRelease, err := ri.AcquireKey(ctx, SpaceKey(dest))
		if err != nil {
			return err
		}
		defer destSRelease()
	}

	// collect all the space cids
	keys, err := ri.cl.HKeys(ctx, sSK).Result()
	if err != nil {
		return
	}
	var cids []string
	for _, k := range keys {
		if strings.HasPrefix(k, "c:") {
			cids = append(cids, k[2:])
		}
	}

	// load entries
	cidEntries, err := ri.CidEntriesByString(ctx, cids)
	if err != nil {
		return
	}
	defer cidEntries.Release()

	// increment cid refs in the dest group
	dGK := GroupKey(dest)
	cmdGroupExists := make([]*redis.IntCmd, len(cidEntries.entries))
	_, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, cidE := range cidEntries.entries {
			cmdGroupExists[i] = pipe.HIncrBy(ctx, dGK, CidKey(cidE.Cid), 1)
		}
		return nil
	})

	// calculate the dest group sums
	for i, cmd := range cmdGroupExists {
		res, _ := cmd.Result()
		if res == 1 {
			destGroup.CidCount++
			destGroup.Size_ += cidEntries.entries[i].Size_
		}
	}

	// remove cids from the src group
	sGK := GroupKey(src)
	var srcCmdGroupDecr = make([]*redis.IntCmd, len(cidEntries.entries))
	_, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, cidE := range cidEntries.entries {
			srcCmdGroupDecr[i] = pipe.HIncrBy(ctx, sGK, CidKey(cidE.Cid), -1)
		}
		return nil
	})
	if err != nil {
		return
	}

	// remove unneeded cids / update sums / save the src group entry
	_, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, cmd := range srcCmdGroupDecr {
			res, _ := cmd.Result()
			if res == 0 {
				srcEntry.group.CidCount--
				srcEntry.group.Size_ -= cidEntries.entries[i].Size_
				pipe.HDel(ctx, sGK, CidKey(cidEntries.entries[i].Cid))
			}
		}
		srcEntry.space.GroupId = dest.GroupId
		srcEntry.space.Save(ctx, src, pipe)
		srcEntry.group.SpaceIds = slices.DeleteFunc(srcEntry.group.SpaceIds, func(spaceId string) bool {
			return spaceId == src.SpaceId
		})
		srcEntry.group.Save(ctx, pipe)
		return nil
	})
	if err != nil {
		return
	}

	_, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		srcEntry.space.Save(ctx, dest, pipe)
		destGroup.AddSpaceId(src.SpaceId)
		destGroup.Save(ctx, pipe)
		return nil
	})

	if sSK != dSK {
		if err = ri.cl.Del(ctx, sSK).Err(); err != nil {
			return
		}
	}
	return
}
