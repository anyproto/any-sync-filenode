package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anyproto/any-sync-filenode/testutil"
)

func TestRedisIndex_SpaceDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		key := newRandKey()

		// space not exists
		ok, err := fx.SpaceDelete(ctx, key)
		require.NoError(t, err)
		assert.False(t, ok)

		// add files
		bs := testutil.NewRandBlocks(5)
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

		groupInfo, err := fx.GroupInfo(ctx, key.GroupId)
		require.NoError(t, err)
		assert.NotEmpty(t, groupInfo.BytesUsage)
		assert.Contains(t, groupInfo.SpaceIds, key.SpaceId)

		require.NoError(t, fx.SetGroupLimit(ctx, key.GroupId, 5000))
		require.NoError(t, fx.SetSpaceLimit(ctx, key, 4000))

		ok, err = fx.SpaceDelete(ctx, key)
		require.NoError(t, err)
		assert.True(t, ok)

		groupInfo, err = fx.GroupInfo(ctx, key.GroupId)
		require.NoError(t, err)
		assert.Empty(t, groupInfo.BytesUsage)
		assert.NotContains(t, groupInfo.SpaceIds, key.SpaceId)
		assert.Equal(t, uint64(5000), groupInfo.AccountLimit)
		assert.Equal(t, uint64(5000), groupInfo.Limit)

		// second call
		ok, err = fx.SpaceDelete(ctx, key)
		require.NoError(t, err)
		assert.False(t, ok)
	})
	t.Run("delete from group", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		key := newRandKey()

		// add files
		bs := testutil.NewRandBlocks(5)
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

		require.NoError(t, fx.FileUnbind(ctx, key, fileId1, fileId2))

		ok, err := fx.MarkSpaceAsDeleted(ctx, key)
		require.NoError(t, err)
		assert.True(t, ok)

		require.NoError(t, fx.cl.Del(ctx, SpaceKey(key)).Err())

		ok, err = fx.SpaceDelete(ctx, key)
		require.NoError(t, err)
		assert.True(t, ok)

		groupInfo, err := fx.GroupInfo(ctx, key.GroupId)
		require.NoError(t, err)
		assert.NotContains(t, groupInfo.SpaceIds, key.SpaceId)
	})
}

func TestRedisIndex_MarkSpaceAsDeleted(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	key := newRandKey()
	ok, err := fx.MarkSpaceAsDeleted(ctx, key)
	require.NoError(t, err)
	assert.True(t, ok)

	err = fx.FileBind(ctx, key, "file", &CidEntries{})
	assert.ErrorIs(t, err, ErrSpaceIsDeleted)
}
