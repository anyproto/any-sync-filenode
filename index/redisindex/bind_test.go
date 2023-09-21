package redisindex

import (
	"github.com/anyproto/any-sync-filenode/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRedisIndex_Bind(t *testing.T) {
	t.Run("first add", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		bs := testutil.NewRandBlocks(3)
		var sumSize uint64
		for _, b := range bs {
			sumSize += uint64(len(b.RawData()))
		}
		spaceId := testutil.NewRandSpaceId()
		fileId := testutil.NewRandCid().String()

		require.NoError(t, fx.Bind(ctx, spaceId, fileId, bs))
		size, err := fx.StorageSize(ctx, spaceId)
		require.NoError(t, err)
		assert.Equal(t, sumSize, size)
	})
	t.Run("two files with same cids", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		bs := testutil.NewRandBlocks(3)
		var sumSize uint64
		for _, b := range bs {
			sumSize += uint64(len(b.RawData()))
		}
		spaceId := testutil.NewRandSpaceId()
		fileId1 := testutil.NewRandCid().String()
		fileId2 := testutil.NewRandCid().String()

		require.NoError(t, fx.Bind(ctx, spaceId, fileId1, bs[:2]))
		require.NoError(t, fx.Bind(ctx, spaceId, fileId2, bs))
		size, err := fx.StorageSize(ctx, spaceId)
		require.NoError(t, err)
		assert.Equal(t, sumSize, size)
	})
	t.Run("bind twice", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		bs := testutil.NewRandBlocks(3)
		var sumSize uint64
		for _, b := range bs {
			sumSize += uint64(len(b.RawData()))
		}
		spaceId := testutil.NewRandSpaceId()
		fileId := testutil.NewRandCid().String()

		require.NoError(t, fx.Bind(ctx, spaceId, fileId, bs))
		require.NoError(t, fx.Bind(ctx, spaceId, fileId, bs))
		size, err := fx.StorageSize(ctx, spaceId)
		require.NoError(t, err)
		assert.Equal(t, sumSize, size)
	})
}

func TestRedisIndex_BindCids(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		bs := testutil.NewRandBlocks(3)
		var sumSize uint64
		for _, b := range bs {
			sumSize += uint64(len(b.RawData()))
		}
		spaceId := testutil.NewRandSpaceId()
		fileId1 := testutil.NewRandCid().String()
		fileId2 := testutil.NewRandCid().String()

		require.NoError(t, fx.Bind(ctx, spaceId, fileId1, bs))
		require.NoError(t, fx.BindCids(ctx, spaceId, fileId2, testutil.BlocksToKeys(bs)))
		size, err := fx.StorageSize(ctx, spaceId)
		require.NoError(t, err)
		assert.Equal(t, sumSize, size)

		fi1, err := fx.FileInfo(ctx, spaceId, fileId1)
		require.NoError(t, err)
		fi2, err := fx.FileInfo(ctx, spaceId, fileId2)
		require.NoError(t, err)

		assert.Equal(t, uint32(len(bs)), fi1.CidCount)
		assert.Equal(t, size, fi1.BytesUsage)
		assert.Equal(t, fi1, fi2)
	})
	t.Run("bind not existing cids", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		bs := testutil.NewRandBlocks(3)
		var sumSize uint64
		for _, b := range bs {
			sumSize += uint64(len(b.RawData()))
		}
		spaceId := testutil.NewRandSpaceId()
		fileId1 := testutil.NewRandCid().String()
		require.Error(t, fx.BindCids(ctx, spaceId, fileId1, testutil.BlocksToKeys(bs)))
	})
}
