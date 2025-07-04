package index

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func (ri *redisIndex) SpaceDelete(ctx context.Context, key Key) (ok bool, err error) {
	entry, release, err := ri.AcquireSpace(ctx, key)
	if err != nil {
		if errors.Is(err, ErrSpaceIsDeleted) {
			return ri.removeSpaceFromGroup(ctx, key)
		}
		return
	}
	defer release()

	return ri.spaceDelete(ctx, key, entry)
}

func (ri *redisIndex) spaceDelete(ctx context.Context, key Key, entry groupSpaceEntry) (ok bool, err error) {
	if !entry.spaceExists {
		return false, nil
	}
	sk := SpaceKey(key)

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

func (ri *redisIndex) MarkSpaceAsDeleted(ctx context.Context, key Key) (ok bool, err error) {
	exists, release, err := ri.AcquireKey(ctx, DelKey(key))
	if err != nil {
		return
	}
	defer release()
	if !exists {
		if err = ri.cl.Set(ctx, DelKey(key), time.Now().Unix(), 0).Err(); err != nil {
			return
		}
		return true, nil
	}
	return false, nil
}

func (ri *redisIndex) removeSpaceFromGroup(ctx context.Context, key Key) (ok bool, err error) {
	gExists, gRelease, err := ri.AcquireKey(ctx, GroupKey(key))
	if err != nil {
		return
	}
	defer gRelease()
	if !gExists {
		return false, nil
	}

	gEntry, err := ri.getGroupEntry(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return
	}
	if !slices.Contains(gEntry.SpaceIds, key.SpaceId) {
		return false, nil
	}
	gEntry.SpaceIds = slices.DeleteFunc(gEntry.SpaceIds, func(spaceId string) bool {
		return spaceId == key.SpaceId
	})
	_, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		gEntry.Save(ctx, pipe)
		return nil
	})
	if err != nil {
		return
	}
	// add spaces broken spaces to the debug list
	_ = ri.cl.LPush(ctx, "debug:spaceIdsToRecheck", key.SpaceId).Err()
	return true, nil
}
