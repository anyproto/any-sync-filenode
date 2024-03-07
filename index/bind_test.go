package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anyproto/any-sync-filenode/testutil"
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
		key := newRandKey()
		fileId := testutil.NewRandCid().String()
		require.NoError(t, fx.BlocksAdd(ctx, bs))
		cids, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		defer cids.Release()

		require.NoError(t, fx.FileBind(ctx, key, fileId, cids))
		fInfo, err := fx.FileInfo(ctx, key, fileId)
		require.NoError(t, err)
		assert.Equal(t, uint64(len(bs)), fInfo[0].CidsCount)
		assert.Equal(t, sumSize, fInfo[0].BytesUsage)
	})

	t.Run("two files with same cids", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		bs := testutil.NewRandBlocks(3)
		var sumSize uint64
		for _, b := range bs {
			sumSize += uint64(len(b.RawData()))
		}
		key := newRandKey()
		fileId1 := testutil.NewRandCid().String()
		fileId2 := testutil.NewRandCid().String()
		require.NoError(t, fx.BlocksAdd(ctx, bs))

		cidsA, err := fx.CidEntriesByBlocks(ctx, bs[:2])
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, key, fileId1, cidsA))
		cidsA.Release()

		cidsB, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		defer cidsB.Release()
		require.NoError(t, fx.FileBind(ctx, key, fileId2, cidsB))

		spaceInfo, err := fx.SpaceInfo(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, SpaceInfo{
			BytesUsage: sumSize,
			FileCount:  2,
			CidsCount:  uint64(len(bs)),
		}, spaceInfo)

		groupInfo, err := fx.GroupInfo(ctx, key.GroupId)
		require.NoError(t, err)
		assert.Equal(t, GroupInfo{
			BytesUsage:   sumSize,
			CidsCount:    uint64(len(bs)),
			AccountLimit: fx.defaultLimit,
			Limit:        fx.defaultLimit,
			SpaceIds:     []string{key.SpaceId},
		}, groupInfo)

	})

	t.Run("bind twice", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		bs := testutil.NewRandBlocks(3)
		var sumSize uint64
		for _, b := range bs {
			sumSize += uint64(len(b.RawData()))
		}
		key := newRandKey()
		fileId := testutil.NewRandCid().String()

		require.NoError(t, fx.BlocksAdd(ctx, bs))
		cids, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		defer cids.Release()

		require.NoError(t, fx.FileBind(ctx, key, fileId, cids))
		require.NoError(t, fx.FileBind(ctx, key, fileId, cids))

		fInfo, err := fx.FileInfo(ctx, key, fileId)
		require.NoError(t, err)
		assert.Equal(t, uint64(len(bs)), fInfo[0].CidsCount)
		assert.Equal(t, sumSize, fInfo[0].BytesUsage)
	})
}
