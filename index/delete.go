package index

import (
	"context"
	"slices"
	"strings"

	"github.com/redis/go-redis/v9"
)

func (ri *redisIndex) SpaceDelete(ctx context.Context, key Key) (ok bool, err error) {
	var (
		gk = groupKey(key)
		sk = spaceKey(key)
	)

	_, gRelease, err := ri.AcquireKey(ctx, gk)
	if err != nil {
		return
	}
	defer gRelease()
	spaceExists, sRelease, err := ri.AcquireKey(ctx, sk)
	if err != nil {
		return
	}
	defer sRelease()

	if !spaceExists {
		return false, nil
	}

	keys, err := ri.cl.HKeys(ctx, sk).Result()
	if err != nil {
		return
	}
	for _, k := range keys {
		if strings.HasPrefix(k, "f:") {
			if err = ri.fileUnbind(ctx, key, k[2:]); err != nil {
				return
			}
		}
	}

	groupInfo, err := ri.getGroupEntry(ctx, key)
	if err != nil {
		return
	}
	if !slices.Contains(groupInfo.SpaceIds, key.SpaceId) {
		return false, nil
	}
	groupInfo.SpaceIds = slices.DeleteFunc(groupInfo.SpaceIds, func(spaceId string) bool {
		return spaceId == key.SpaceId
	})

	_, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		groupInfo.Save(ctx, key, ri.cl)
		return nil
	})
	if err != nil {
		return
	}
	return true, nil
}
