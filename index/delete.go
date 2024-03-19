package index

import (
	"context"
	"slices"
	"strings"

	"github.com/redis/go-redis/v9"
)

func (ri *redisIndex) SpaceDelete(ctx context.Context, key Key) (ok bool, err error) {
	entry, release, err := ri.AcquireSpace(ctx, key)
	if err != nil {
		return
	}
	defer release()

	return ri.spaceDelete(ctx, key, entry)
}

func (ri *redisIndex) spaceDelete(ctx context.Context, key Key, entry groupSpaceEntry) (ok bool, err error) {
	if !entry.spaceExists {
		return false, nil
	}
	sk := spaceKey(key)

	keys, err := ri.cl.HKeys(ctx, sk).Result()
	if err != nil {
		return
	}
	for _, k := range keys {
		if strings.HasPrefix(k, "f:") {
			if err = ri.fileUnbind(ctx, key, entry, k[2:]); err != nil {
				return
			}
		}
	}

	if !slices.Contains(entry.group.SpaceIds, key.SpaceId) {
		return false, nil
	}
	entry.group.SpaceIds = slices.DeleteFunc(entry.group.SpaceIds, func(spaceId string) bool {
		return spaceId == key.SpaceId
	})

	// in case of isolated space - return limit to the group
	if entry.space.Limit != 0 {
		entry.group.Limit += entry.space.Limit
	}

	_, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		entry.group.Save(ctx, pipe)
		pipe.Del(ctx, sk)
		return nil
	})
	if err != nil {
		return
	}
	return true, nil
}
