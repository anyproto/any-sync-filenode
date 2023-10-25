package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anyproto/any-sync-filenode/testutil"
)

func TestRedisIndex_UnBind(t *testing.T) {
	t.Run("unbind non existent", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		key := newRandKey()
		require.NoError(t, fx.FileUnbind(ctx, key, testutil.NewRandCid().String()))
	})
	t.Run("unbind single", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		key := newRandKey()
		bs := testutil.NewRandBlocks(5)

		require.NoError(t, fx.BlocksAdd(ctx, bs))
		cids, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)

		fileId := testutil.NewRandCid().String()
		require.NoError(t, fx.FileBind(ctx, key, fileId, cids))
		cids.Release()

		require.NoError(t, fx.FileUnbind(ctx, key, fileId))

		groupInfo, err := fx.GroupInfo(ctx, key.GroupId)
		require.NoError(t, err)
		assert.Empty(t, groupInfo.CidsCount)
		assert.Empty(t, groupInfo.BytesUsage)
		spaceInfo, err := fx.SpaceInfo(ctx, key)
		require.NoError(t, err)
		assert.Empty(t, spaceInfo.FileCount)
		assert.Empty(t, spaceInfo.BytesUsage)

		cids, err = fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		defer cids.Release()
		for _, r := range cids.entries {
			assert.Equal(t, int32(0), r.Refs)
		}
	})
	t.Run("unbind intersection file", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		key := newRandKey()
		bs := testutil.NewRandBlocks(5)
		var file1Size = uint64(len(bs[0].RawData()) + len(bs[1].RawData()))
		require.NoError(t, fx.BlocksAdd(ctx, bs))
		cids, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		cids.Release()

		// bind to files with intersected cids
		fileId1 := testutil.NewRandCid().String()
		fileId2 := testutil.NewRandCid().String()

		cids1, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, key, fileId1, cids1))
		cids1.Release()

		cids2, err := fx.CidEntriesByBlocks(ctx, bs[:2])
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, key, fileId2, cids2))
		cids2.Release()

		// remove file1
		require.NoError(t, fx.FileUnbind(ctx, key, fileId1))

		groupInfo, err := fx.GroupInfo(ctx, key.GroupId)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), groupInfo.CidsCount)
		assert.Equal(t, file1Size, groupInfo.BytesUsage)
		spaceInfo, err := fx.SpaceInfo(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, uint32(1), spaceInfo.FileCount)
		assert.Equal(t, file1Size, spaceInfo.BytesUsage)

		cids, err = fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		defer cids.Release()

	})
}
