package index

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-sync-filenode/config"
	"github.com/anyproto/any-sync-filenode/index/indexproto"
	"github.com/anyproto/any-sync-filenode/testutil"
)

func TestRedisIndex_BlocksAdd(t *testing.T) {
	bs := testutil.NewRandBlocks(5)
	fx := newFixture(t)
	defer fx.Finish(t)

	require.NoError(t, fx.BlocksAdd(ctx, bs))

	result, err := fx.CidEntriesByBlocks(ctx, bs)
	require.NoError(t, err)
	defer result.Release()

	require.Len(t, result.entries, len(bs))
	for _, e := range result.entries {
		assert.NotEmpty(t, e.Size)
		assert.NotEmpty(t, e.CreateTime)
		assert.NotEmpty(t, e.UpdateTime)
		assert.NotEmpty(t, e.Version)
	}
}

func TestRedisIndex_CidEntries(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bs := testutil.NewRandBlocks(5)
		fx := newFixture(t)
		defer fx.Finish(t)

		require.NoError(t, fx.BlocksAdd(ctx, bs))

		cids := testutil.BlocksToKeys(bs)

		result, err := fx.CidEntries(ctx, cids)
		defer result.Release()
		require.NoError(t, err)
		require.Len(t, result.entries, len(bs))
	})
	t.Run("not all cids", func(t *testing.T) {
		bs := testutil.NewRandBlocks(5)
		fx := newFixture(t)
		defer fx.Finish(t)

		require.NoError(t, fx.BlocksAdd(ctx, bs[:3]))

		cids := testutil.BlocksToKeys(bs)
		fx.persistStore.EXPECT().Get(ctx, gomock.Any()).Return(nil, fmt.Errorf("err")).AnyTimes()
		_, err := fx.CidEntries(ctx, cids)
		assert.EqualError(t, err, ErrCidsNotExist.Error())
	})
	t.Run("migrate old cids", func(t *testing.T) {
		bs := testutil.NewRandBlocks(5)
		fx := newFixture(t)
		defer fx.Finish(t)

		for _, b := range bs {
			// save old entry, without version
			entry := &cidEntry{
				Cid: b.Cid(),
				CidEntry: &indexproto.CidEntry{
					Size:       uint64(len(b.RawData())),
					CreateTime: 1,
					UpdateTime: 2,
				},
			}
			require.NoError(t, entry.Save(ctx, fx.cl))
		}

		cids := testutil.BlocksToKeys(bs)

		result, err := fx.CidEntries(ctx, cids)
		defer result.Release()
		require.NoError(t, err)
		require.Len(t, result.entries, len(bs))
		for _, e := range result.entries {
			assert.NotEmpty(t, e.Size)
			assert.NotEmpty(t, e.CreateTime)
			assert.NotEmpty(t, e.UpdateTime)
			assert.NotEmpty(t, e.Version)
		}
	})
	t.Run("restore from store", func(t *testing.T) {
		bs := testutil.NewRandBlocks(4)
		fx := newFixture(t)
		defer fx.Finish(t)

		require.NoError(t, fx.BlocksAdd(ctx, bs[:3]))

		cids := testutil.BlocksToKeys(bs)

		fx.persistStore.EXPECT().Get(ctx, bs[3].Cid()).Return(bs[3], nil)

		result, err := fx.CidEntries(ctx, cids)
		defer result.Release()
		require.NoError(t, err)
		require.Len(t, result.entries, len(bs))
		t.Log(result.entries[3])
	})
}

func TestRedisIndex_CidExistsInSpace(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	key := newRandKey()

	bs := testutil.NewRandBlocks(5)
	require.NoError(t, fx.BlocksAdd(ctx, bs))

	cids, err := fx.CidEntriesByBlocks(ctx, bs[:2])
	require.NoError(t, err)
	require.NoError(t, fx.FileBind(ctx, key, "fileId", cids))
	cids.Release()

	exists, err := fx.CidExistsInSpace(ctx, key, testutil.BlocksToKeys(bs))
	require.NoError(t, err)

	require.Len(t, exists, 2)
	assert.Equal(t, testutil.BlocksToKeys(bs[:2]), exists)
}

func TestRedisIndex_CidExists(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	bs := testutil.NewRandBlocks(5)
	require.NoError(t, fx.BlocksAdd(ctx, bs[:2]))

	for i, b := range bs {
		ok, err := fx.CidExists(ctx, b.Cid())
		require.NoError(t, err)
		if i < 2 {
			assert.True(t, ok)
		} else {
			assert.False(t, ok)
		}
	}
}

func TestRedisIndex_CidExistsBulk(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)

		res, err := fx.CidExistsBulk(ctx, nil)
		require.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("redis hits and misses", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)

		live := testutil.NewRandBlocks(3)
		missing := testutil.NewRandBlocks(2)
		require.NoError(t, fx.BlocksAdd(ctx, live))

		all := append(append([]cid.Cid{}, testutil.BlocksToKeys(live)...), testutil.BlocksToKeys(missing)...)

		res, err := fx.CidExistsBulk(ctx, all)
		require.NoError(t, err)
		require.Equal(t, []bool{true, true, true, false, false}, res)
	})

	t.Run("hit via persistStore after eviction", func(t *testing.T) {
		fx := newFixtureConfig(t, &config.Config{DefaultLimit: 1024, PersistTtl: 1})
		defer fx.Finish(t)

		bs := testutil.NewRandBlocks(2)
		require.NoError(t, fx.BlocksAdd(ctx, bs))

		dumps := make(map[string][]byte, len(bs))
		for _, b := range bs {
			b := b
			fx.persistStore.EXPECT().IndexPut(ctx, CidKey(b.Cid()), gomock.Any()).
				Do(func(_ context.Context, key string, value []byte) {
					dumps[key] = append([]byte(nil), value...)
				})
		}

		time.Sleep(time.Second * 2)
		fx.PersistKeys(ctx)

		for _, b := range bs {
			b := b
			fx.persistStore.EXPECT().IndexGet(gomock.Any(), CidKey(b.Cid())).
				DoAndReturn(func(_ context.Context, key string) ([]byte, error) {
					return dumps[key], nil
				})
		}

		res, err := fx.CidExistsBulk(ctx, testutil.BlocksToKeys(bs))
		require.NoError(t, err)
		assert.Equal(t, []bool{true, true}, res)
	})
}
