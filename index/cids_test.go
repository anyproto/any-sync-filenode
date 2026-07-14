package index

import (
	"fmt"
	"testing"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

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

func TestRedisIndex_DeleteUnboundCid(t *testing.T) {
	t.Run("unbound cid deleted", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		b := testutil.NewRandBlock(1024)
		c := b.Cid()
		require.NoError(t, fx.BlocksAdd(ctx, []blocks.Block{b}))

		countBefore, err := fx.cl.Get(ctx, cidCount).Int64()
		require.NoError(t, err)
		sizeBefore, err := fx.cl.Get(ctx, cidSizeSumKey).Int64()
		require.NoError(t, err)

		fx.persistStore.EXPECT().DeleteMany(gomock.Any(), []cid.Cid{c}).Return(nil)
		fx.persistStore.EXPECT().IndexDelete(gomock.Any(), CidKey(c)).Return(nil)

		ok, err := fx.DeleteUnboundCid(ctx, c)
		require.NoError(t, err)
		assert.True(t, ok)

		exists, err := fx.CidExists(ctx, c)
		require.NoError(t, err)
		assert.False(t, exists)

		countAfter, err := fx.cl.Get(ctx, cidCount).Int64()
		require.NoError(t, err)
		assert.Equal(t, countBefore-1, countAfter)
		sizeAfter, err := fx.cl.Get(ctx, cidSizeSumKey).Int64()
		require.NoError(t, err)
		assert.Equal(t, sizeBefore-int64(len(b.RawData())), sizeAfter)
	})

	t.Run("bound cid refused", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		b := testutil.NewRandBlock(1024)
		key := newRandKey()
		fileId := testutil.NewRandCid().String()
		require.NoError(t, fx.BlocksAdd(ctx, []blocks.Block{b}))
		cids, err := fx.CidEntriesByBlocks(ctx, []blocks.Block{b})
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, key, fileId, cids))
		cids.Release()

		_, err = fx.DeleteUnboundCid(ctx, b.Cid())
		require.ErrorIs(t, err, ErrCidIsBound)

		exists, err := fx.CidExists(ctx, b.Cid())
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existent cid is no-op", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)

		ok, err := fx.DeleteUnboundCid(ctx, testutil.NewRandCid())
		require.NoError(t, err)
		assert.False(t, ok)
	})
}
