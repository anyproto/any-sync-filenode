package deletelog

import (
	"context"
	"errors"
	"time"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/commonfile/fileblockstore"
	"github.com/anyproto/any-sync/coordinator/coordinatorclient"
	"github.com/anyproto/any-sync/coordinator/coordinatorproto"
	"github.com/anyproto/any-sync/util/periodicsync"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/anyproto/any-sync-filenode/index"
	"github.com/anyproto/any-sync-filenode/redisprovider"
	"github.com/anyproto/any-sync-filenode/store"
)

const CName = "filenode.deletionLog"

const lastKey = "deletionLastId.{system}"

const recordsLimit = 1000

var log = logger.NewNamed(CName)

func New() app.ComponentRunnable {
	return new(deleteLog)
}

type deleteLog struct {
	redis             redis.UniversalClient
	coordinatorClient coordinatorclient.CoordinatorClient
	redsync           *redsync.Redsync
	ticker            periodicsync.PeriodicSync
	index             index.Index
	store             store.Store
	disableTicker     bool
}

func (d *deleteLog) Init(a *app.App) (err error) {
	d.redis = a.MustComponent(redisprovider.CName).(redisprovider.RedisProvider).Redis()
	d.coordinatorClient = a.MustComponent(coordinatorclient.CName).(coordinatorclient.CoordinatorClient)
	d.redsync = redsync.New(goredis.NewPool(d.redis))
	d.index = a.MustComponent(index.CName).(index.Index)
	d.store = a.MustComponent(fileblockstore.CName).(store.Store)
	return
}

func (d *deleteLog) Name() (name string) {
	return CName
}

func (d *deleteLog) Run(ctx context.Context) (err error) {
	if !d.disableTicker {
		d.ticker = periodicsync.NewPeriodicSync(30, time.Hour*2, d.checkLog, log)
		d.ticker.Run()
	}
	return
}

func (d *deleteLog) checkLog(ctx context.Context) (err error) {
	mu := d.redsync.NewMutex("_lock:deletion", redsync.WithExpiry(time.Hour*2))
	if err = mu.LockContext(ctx); err != nil {
		return
	}
	defer func() {
		_, _ = mu.Unlock()
	}()
	st := time.Now()
	lastId, err := d.redis.Get(ctx, lastKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return
	}

	recs, err := d.coordinatorClient.DeletionLog(ctx, lastId, recordsLimit)
	if err != nil {
		return
	}
	var handledCount, deletedCount int
	for _, rec := range recs {
		if rec.Status == coordinatorproto.DeletionLogRecordStatus_Remove && rec.FileGroup != "" {
			key := index.Key{
				GroupId: rec.FileGroup,
				SpaceId: rec.SpaceId,
			}
			cids, err := d.index.SpaceDelete(ctx, key)
			if err != nil && !errors.Is(err, redis.Nil) && !errors.Is(err, index.ErrSpaceIsDeleted) {
				return err
			}
			if len(cids) > 0 {
				if err := d.store.DeleteMany(ctx, cids); err != nil {
					log.WarnCtx(ctx, "failed to delete cids from store", zap.Error(err))
				}
			}
			handledCount++
			if _, err = d.index.MarkSpaceAsDeleted(ctx, key); err != nil {
				return err
			}
			deletedCount++
		}
		if err = d.redis.Set(ctx, lastKey, rec.Id, 0).Err(); err != nil {
			return err
		}
	}
	log.Info("processing deletion log",
		zap.Int("records", len(recs)),
		zap.Int("handled", handledCount),
		zap.Int("deleted", deletedCount),
		zap.Duration("dur", time.Since(st)),
	)
	return
}

func (d *deleteLog) Close(ctx context.Context) (err error) {
	if d.ticker != nil {
		d.ticker.Close()
	}
	return
}
