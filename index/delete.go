package index

import (
	"context"
	"slices"
	"strings"

	"github.com/redis/go-redis/v9"
)

func (ri *redisIndex) SpaceDelete(ctx context.Context, key Key) (ok bool, err error) {
	var (
		sk = spaceKey(key)
	)

	entry, release, err := ri.AcquireSpace(ctx, key)
	if err != nil {
		return
	}
	defer release()

	if !entry.spaceExists {
		return false, nil
	}

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

	_, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		entry.group.Save(ctx, key, pipe)
		return nil
	})
	if err != nil {
		return
	}
	return true, nil
}
