package redisindex

import (
	"github.com/anyproto/any-sync-filenode/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRedisIndex_UnBind(t *testing.T) {
	t.Run("unbind non existent", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		spaceId := testutil.NewRandSpaceId()
		require.NoError(t, fx.UnBind(ctx, spaceId, testutil.NewRandCid().String()))
	})
	t.Run("unbind single", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		spaceId := testutil.NewRandSpaceId()
		bs := testutil.NewRandBlocks(5)
		fileId := testutil.NewRandCid().String()
		require.NoError(t, fx.Bind(ctx, spaceId, fileId, bs))

		require.NoError(t, fx.UnBind(ctx, spaceId, fileId))
		size, err := fx.StorageSize(ctx, spaceId)
		require.NoError(t, err)
		assert.Empty(t, size)

		var cidsB [][]byte
		for _, b := range bs {
			cidsB = append(cidsB, b.Cid().Bytes())
		}

		res, err := fx.cidInfoByByteKeys(ctx, cidsB)
		require.NoError(t, err)
		for _, r := range res {
			assert.Equal(t, int32(0), r.Refs)
		}
	})
	t.Run("unbind two files", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		spaceId := testutil.NewRandSpaceId()
		bs := testutil.NewRandBlocks(5)
		fileId1 := testutil.NewRandCid().String()
		fileId2 := testutil.NewRandCid().String()
		require.NoError(t, fx.Bind(ctx, spaceId, fileId1, bs))
		require.NoError(t, fx.Bind(ctx, spaceId, fileId2, bs[:3]))

		si, err := fx.StorageInfo(ctx, spaceId)
		require.NoError(t, err)
		assert.Equal(t, 2, si.FileCount)
		assert.Equal(t, len(bs), si.CidCount)

		require.NoError(t, fx.UnBind(ctx, spaceId, fileId1))
		si, err = fx.StorageInfo(ctx, spaceId)
		require.NoError(t, err)
		assert.Equal(t, 1, si.FileCount)
		assert.Equal(t, 3, si.CidCount)
	})
}
