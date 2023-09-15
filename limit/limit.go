//go:generate mockgen -destination mock_limit/mock_limit.go github.com/anyproto/any-sync-filenode/limit Limit
package limit

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/ocache"
	"github.com/anyproto/any-sync/coordinator/coordinatorclient"
	"github.com/anyproto/any-sync/net/peer"
	"go.uber.org/atomic"
	"strings"
	"time"
)

const CName = "filenode.limit"

func New() Limit {
	return new(limit)
}

type Limit interface {
	Check(ctx context.Context, spaceId string) (limit uint64, storageKey string, err error)
	app.ComponentRunnable
}

type limit struct {
	cl    coordinatorclient.CoordinatorClient
	cache ocache.OCache
}

func (l *limit) Init(a *app.App) (err error) {
	l.cl = a.MustComponent(coordinatorclient.CName).(coordinatorclient.CoordinatorClient)
	l.cache = ocache.New(l.fetchLimit,
		ocache.WithTTL(time.Minute*10),
		ocache.WithGCPeriod(time.Minute*5),
	)
	return nil
}

func (l *limit) Name() (name string) {
	return CName
}

func (l *limit) Run(ctx context.Context) (err error) {
	return
}

func (l *limit) Check(ctx context.Context, spaceId string) (limit uint64, storageKey string, err error) {
	identity, err := peer.CtxIdentity(ctx)
	if err != nil {
		return
	}
	obj, err := l.cache.Get(ctx, encodeId(spaceId, identity))
	if err != nil {
		return
	}
	e := obj.(*entry)
	return e.limit, e.storageKey, nil
}

func (l *limit) fetchLimit(ctx context.Context, id string) (value ocache.Object, err error) {
	spaceId, identity, err := decodeId(id)
	if err != nil {
		return
	}
	result, err := l.cl.FileLimitCheck(ctx, spaceId, identity)
	if err != nil {
		return nil, err
	}

	if result.StorageKey == "" {
		result.StorageKey = spaceId
	}
	return &entry{
		spaceId:    spaceId,
		identity:   identity,
		lastUsage:  atomic.NewTime(time.Now()),
		limit:      result.Limit,
		storageKey: result.StorageKey,
	}, nil
}

func (l *limit) Close(ctx context.Context) (err error) {
	return l.cache.Close()
}

func encodeId(spaceId string, identity []byte) string {
	return spaceId + ":" + hex.EncodeToString(identity)
}

func decodeId(id string) (spaceId string, identity []byte, err error) {
	idx := strings.Index(id, ":")
	if idx == -1 {
		err = fmt.Errorf("unexpected limit id")
	}
	spaceId = id[:idx]
	if identity, err = hex.DecodeString(id[idx+1:]); err != nil {
		return
	}
	return
}

type entry struct {
	spaceId    string
	identity   []byte
	lastUsage  *atomic.Time
	limit      uint64
	storageKey string
}

func (e *entry) TryClose(objectTTL time.Duration) (res bool, err error) {
	if time.Now().Sub(e.lastUsage.Load()) < objectTTL {
		return false, nil
	}
	return true, e.Close()
}

func (e *entry) Close() error {
	return nil
}
