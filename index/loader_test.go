package index

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	blocks "github.com/ipfs/go-block-format"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-sync-filenode/config"
	"github.com/anyproto/any-sync-filenode/testutil"
)

func TestRedisIndex_PersistKeys(t *testing.T) {
	t.Run("no keys", func(t *testing.T) {
		fx := newFixtureConfig(t, &config.Config{PersistTtl: 2})
		defer fx.Finish(t)
		bs := testutil.NewRandBlocks(5)
		require.NoError(t, fx.BlocksAdd(ctx, bs))
		fx.PersistKeys(ctx)
	})
	t.Run("delete", func(t *testing.T) {
		fx := newFixtureConfig(t, &config.Config{PersistTtl: 1})
		defer fx.Finish(t)
		bs := testutil.NewRandBlocks(5)
		for _, b := range bs {
			_, release, _ := fx.AcquireKey(ctx, CidKey(b.Cid()))
			release()
			fx.persistStore.EXPECT().IndexDelete(ctx, CidKey(b.Cid()))
		}
		time.Sleep(time.Second * 3)
		fx.PersistKeys(ctx)
	})
	t.Run("persist", func(t *testing.T) {
		fx := newFixtureConfig(t, &config.Config{PersistTtl: 1})
		defer fx.Finish(t)
		bs := testutil.NewRandBlocks(5)
		require.NoError(t, fx.BlocksAdd(ctx, bs))
		for _, b := range bs {
			fx.persistStore.EXPECT().IndexPut(ctx, CidKey(b.Cid()), gomock.Any())
		}

		time.Sleep(time.Second * 3)
		bs2 := testutil.NewRandBlocks(5)
		require.NoError(t, fx.BlocksAdd(ctx, bs2))

		fx.PersistKeys(ctx)

		for _, b := range bs {
			res, err := fx.cl.BFExists(ctx, bloomFilterKey(CidKey(b.Cid())), CidKey(b.Cid())).Result()
			require.NoError(t, err)
			assert.True(t, res)
		}
	})
	t.Run("bloom filter tx error", func(t *testing.T) {
		hook := &failBFAddHook{}
		fx := newFixtureConfig(t, &config.Config{PersistTtl: 1}, hook)
		defer fx.Finish(t)

		b := testutil.NewRandBlock(10)
		k := CidKey(b.Cid())
		require.NoError(t, fx.BlocksAdd(ctx, []blocks.Block{b}))
		time.Sleep(time.Second * 3)

		// the tx fails: the key has to stay in redis and in the queue
		hook.fail.Store(true)
		fx.persistStore.EXPECT().IndexPut(ctx, k, gomock.Any())
		fx.PersistKeys(ctx)

		ex, err := fx.cl.Exists(ctx, k).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), ex)
		inBloom, err := fx.cl.BFExists(ctx, bloomFilterKey(k), k).Result()
		require.NoError(t, err)
		assert.False(t, inBloom)

		// the next cycle persists the key again instead of deleting it from the persistent store
		hook.fail.Store(false)
		var persisted []byte
		fx.persistStore.EXPECT().IndexPut(ctx, k, gomock.Any()).Do(func(_ context.Context, _ string, value []byte) {
			persisted = value
		})
		fx.PersistKeys(ctx)

		ex, err = fx.cl.Exists(ctx, k).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), ex)
		inBloom, err = fx.cl.BFExists(ctx, bloomFilterKey(k), k).Result()
		require.NoError(t, err)
		assert.True(t, inBloom)

		fx.persistStore.EXPECT().IndexGet(ctx, k).Return(persisted, nil)
		exists, err := fx.CidExists(ctx, b.Cid())
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

// failBFAddHook fails a pipeline that contains BF.ADD before it is sent, like a broken connection would
type failBFAddHook struct {
	fail atomic.Bool
}

func (h *failBFAddHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h *failBFAddHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		return next(ctx, cmd)
	}
}

func (h *failBFAddHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		if h.fail.Load() {
			for _, cmd := range cmds {
				if cmd.Name() == "bf.add" {
					return errors.New("injected tx error")
				}
			}
		}
		return next(ctx, cmds)
	}
}

func TestRedisIndex_AcquireKey(t *testing.T) {
	fx := newFixtureConfig(t, &config.Config{PersistTtl: 1})
	defer fx.Finish(t)

	bs := testutil.NewRandBlocks(5)
	require.NoError(t, fx.BlocksAdd(ctx, bs))
	for _, b := range bs {
		fx.persistStore.EXPECT().IndexPut(ctx, CidKey(b.Cid()), gomock.Any()).Do(func(_ context.Context, key string, value []byte) {
			if key == CidKey(bs[0].Cid()) {
				fx.persistStore.EXPECT().IndexGet(ctx, key).Return(nil, nil)
			} else {
				fx.persistStore.EXPECT().IndexGet(ctx, key).Return(value, nil)
			}
		})
	}
	time.Sleep(time.Second * 3)
	fx.PersistKeys(ctx)

	for i, b := range bs {
		ex, release, err := fx.AcquireKey(ctx, CidKey(b.Cid()))
		require.NoError(t, err)
		if i == 0 {
			require.False(t, ex)
		} else {
			require.True(t, ex)
		}
		release()
	}

}
