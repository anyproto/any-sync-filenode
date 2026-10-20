package index

import (
	"context"
	"errors"
	"net"
	"slices"
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
		for _, tc := range []struct {
			name    string
			inject  func(fx *fixture, hook *failBFAddHook, k string)
			recover func(fx *fixture, hook *failBFAddHook, k string)
		}{
			{
				name: "tx not executed",
				inject: func(fx *fixture, hook *failBFAddHook, k string) {
					hook.mode.Store(failBeforeExec)
				},
				recover: func(fx *fixture, hook *failBFAddHook, k string) {
					hook.mode.Store(0)
				},
			},
			{
				// MULTI/EXEC doesn't roll back: ZREM is applied
				name: "BF.ADD fails in exec",
				inject: func(fx *fixture, hook *failBFAddHook, k string) {
					require.NoError(t, fx.cl.Set(ctx, bloomFilterKey(k), "not a filter", 0).Err())
				},
				recover: func(fx *fixture, hook *failBFAddHook, k string) {
					require.NoError(t, fx.cl.Del(ctx, bloomFilterKey(k)).Err())
				},
			},
			{
				// ZREM and BF.ADD are applied, the client gets an error
				name: "reply lost",
				inject: func(fx *fixture, hook *failBFAddHook, k string) {
					hook.mode.Store(failAfterExec)
				},
				recover: func(fx *fixture, hook *failBFAddHook, k string) {
					hook.mode.Store(0)
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				hook := &failBFAddHook{}
				fx := newFixtureConfig(t, &config.Config{PersistTtl: 1}, hook)
				defer fx.Finish(t)

				b := testutil.NewRandBlock(10)
				k := CidKey(b.Cid())
				require.NoError(t, fx.BlocksAdd(ctx, []blocks.Block{b}))
				score, err := fx.cl.ZScore(ctx, storeKey(k), k).Result()
				require.NoError(t, err)
				time.Sleep(time.Second * 3)

				// the tx fails: the key has to stay in redis and in the queue with its score.
				// No IndexDelete call is expected
				tc.inject(fx, hook, k)
				fx.persistStore.EXPECT().IndexPut(ctx, k, gomock.Any())
				fx.PersistKeys(ctx)

				ex, err := fx.cl.Exists(ctx, k).Result()
				require.NoError(t, err)
				assert.Equal(t, int64(1), ex)
				queuedScore, err := fx.cl.ZScore(ctx, storeKey(k), k).Result()
				require.NoError(t, err)
				assert.Equal(t, score, queuedScore)

				// the next cycle persists the key again
				tc.recover(fx, hook, k)
				var persisted []byte
				fx.persistStore.EXPECT().IndexPut(ctx, k, gomock.Any()).Do(func(_ context.Context, _ string, value []byte) {
					persisted = value
				})
				fx.PersistKeys(ctx)

				ex, err = fx.cl.Exists(ctx, k).Result()
				require.NoError(t, err)
				assert.Equal(t, int64(0), ex)
				inBloom, err := fx.cl.BFExists(ctx, bloomFilterKey(k), k).Result()
				require.NoError(t, err)
				assert.True(t, inBloom)
				_, err = fx.cl.ZScore(ctx, storeKey(k), k).Result()
				assert.ErrorIs(t, err, redis.Nil)

				fx.persistStore.EXPECT().IndexGet(ctx, k).Return(persisted, nil)
				exists, err := fx.CidExists(ctx, b.Cid())
				require.NoError(t, err)
				assert.True(t, exists)
			})
		}
	})
}

const (
	failBeforeExec int32 = iota + 1
	failAfterExec
)

// failBFAddHook fails a pipeline that contains BF.ADD: before it is sent, like a broken connection,
// or after it is executed, like a lost reply
type failBFAddHook struct {
	mode atomic.Int32
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
		mode := h.mode.Load()
		if mode == 0 || !slices.ContainsFunc(cmds, func(cmd redis.Cmder) bool { return cmd.Name() == "bf.add" }) {
			return next(ctx, cmds)
		}
		if mode == failAfterExec {
			if err := next(ctx, cmds); err != nil {
				return err
			}
		}
		return errors.New("injected tx error")
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

	// all the keys were in the bloom filter, the first one was missing in the persistent store
	assert.Equal(t, uint64(len(bs)), fx.bloomStat.positive.Load())
	assert.Equal(t, uint64(1), fx.bloomStat.falsePositive.Load())

	// restored keys are in redis: no bloom filter lookups
	for _, b := range bs[1:] {
		ex, err := fx.CheckKey(ctx, CidKey(b.Cid()))
		require.NoError(t, err)
		require.True(t, ex)
	}
	// absent keys that are not in the bloom filter don't count
	ex, err := fx.CheckKey(ctx, CidKey(testutil.NewRandCid()))
	require.NoError(t, err)
	require.False(t, ex)
	assert.Equal(t, uint64(len(bs)), fx.bloomStat.positive.Load())
	assert.Equal(t, uint64(1), fx.bloomStat.falsePositive.Load())
}

func TestRedisIndex_AcquireSpace(t *testing.T) {
	inPersistQueue := func(fx *fixture, key string) bool {
		_, err := fx.cl.ZScore(ctx, storeKey(key), key).Result()
		if errors.Is(err, redis.Nil) {
			return false
		}
		require.NoError(t, err)
		return true
	}

	fx := newFixture(t)
	defer fx.Finish(t)
	key := newRandKey()

	entry, release, err := fx.AcquireSpace(ctx, key)
	require.NoError(t, err)
	release()
	assert.False(t, entry.spaceExists)
	// the space and group keys are created by the caller, so they are queued even if absent
	assert.True(t, inPersistQueue(fx, SpaceKey(key)))
	assert.True(t, inPersistQueue(fx, GroupKey(key)))
	// the del key is only checked, an absent one must not be queued
	assert.False(t, inPersistQueue(fx, DelKey(key)))

	ok, err := fx.MarkSpaceAsDeleted(ctx, key)
	require.NoError(t, err)
	require.True(t, ok)
	assert.True(t, inPersistQueue(fx, DelKey(key)))

	_, _, err = fx.AcquireSpace(ctx, key)
	assert.ErrorIs(t, err, ErrSpaceIsDeleted)
}
