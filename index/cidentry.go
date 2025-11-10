package index

import (
	"context"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/redis/go-redis/v9"

	"github.com/anyproto/any-sync-filenode/index/indexproto"
)

type CidEntries struct {
	entries []*cidEntry
	mu      sync.Mutex
}

func (ce *CidEntries) Release() {
	for _, entry := range ce.entries {
		if entry.release != nil {
			entry.release()
		}
	}
	return
}

func (ce *CidEntries) Add(entry *cidEntry) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.entries = append(ce.entries, entry)
}

type cidEntry struct {
	Cid     cid.Cid
	release func()
	*indexproto.CidEntry
}

func (ce *cidEntry) Save(ctx context.Context, cl redis.Cmdable) error {
	ce.UpdateTime = time.Now().Unix()
	data, err := ce.MarshalVT()
	if err != nil {
		return err
	}
	return cl.Set(ctx, CidKey(ce.Cid), data, 0).Err()
}
