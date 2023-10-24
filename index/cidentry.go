package index

import (
	"context"
	"slices"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/redis/go-redis/v9"

	"github.com/anyproto/any-sync-filenode/index/indexproto"
)

type CidEntries struct {
	entries []*cidEntry
}

func (ce *CidEntries) Release() {
	for _, entry := range ce.entries {
		if entry.release != nil {
			entry.release()
		}
	}
	return
}

type cidEntry struct {
	Cid     cid.Cid
	release func()
	*indexproto.CidEntry
}

func (ce *cidEntry) AddGroupId(groupId string) {
	if !slices.Contains(ce.GroupIds, groupId) {
		ce.GroupIds = append(ce.GroupIds, groupId)
	}
}

func (ce *cidEntry) RemoveGroupId(id string) {
	ce.GroupIds = slices.DeleteFunc(ce.GroupIds, func(s string) bool {
		return s == id
	})
}

func (ce *cidEntry) Save(ctx context.Context, cl redis.Cmdable) error {
	ce.UpdateTime = time.Now().Unix()
	data, err := ce.Marshal()
	if err != nil {
		return err
	}
	return cl.Set(ctx, cidKey(ce.Cid), data, 0).Err()
}
