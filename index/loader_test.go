package index

import (
	"context"
	"testing"
	"time"

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
	t.Run("persist", func(t *testing.T) {
		fx := newFixtureConfig(t, &config.Config{PersistTtl: 1})
		defer fx.Finish(t)
		bs := testutil.NewRandBlocks(5)
		require.NoError(t, fx.BlocksAdd(ctx, bs))
		for _, b := range bs {
			fx.persistStore.EXPECT().IndexPut(ctx, cidKey(b.Cid()), gomock.Any())
		}

		time.Sleep(time.Second * 3)
		bs2 := testutil.NewRandBlocks(5)
		require.NoError(t, fx.BlocksAdd(ctx, bs2))

		fx.PersistKeys(ctx)

		for _, b := range bs {
			res, err := fx.cl.BFExists(ctx, bloomFilterKey(cidKey(b.Cid())), cidKey(b.Cid())).Result()
			require.NoError(t, err)
			assert.True(t, res)
		}
	})
}

func TestRedisIndex_AcquireKey(t *testing.T) {
	fx := newFixtureConfig(t, &config.Config{PersistTtl: 1})
	defer fx.Finish(t)

	bs := testutil.NewRandBlocks(5)
	require.NoError(t, fx.BlocksAdd(ctx, bs))
	for _, b := range bs {
		fx.persistStore.EXPECT().IndexPut(ctx, cidKey(b.Cid()), gomock.Any()).Do(func(_ context.Context, key string, value []byte) {
			if key == cidKey(bs[0].Cid()) {
				fx.persistStore.EXPECT().IndexGet(ctx, key).Return(nil, nil)
			} else {
				fx.persistStore.EXPECT().IndexGet(ctx, key).Return(value, nil)
			}
		})
	}
	time.Sleep(time.Second * 3)
	fx.PersistKeys(ctx)

	for i, b := range bs {
		ex, release, err := fx.AcquireKey(ctx, cidKey(b.Cid()))
		require.NoError(t, err)
		if i == 0 {
			require.False(t, ex)
		} else {
			require.True(t, ex)
		}
		release()
	}

}
