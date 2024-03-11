package index

import (
	"testing"

	"github.com/anyproto/any-sync/commonfile/fileproto/fileprotoerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anyproto/any-sync-filenode/testutil"
)

func TestRedisIndex_CheckLimits(t *testing.T) {
	t.Run("isolated space", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)

		k := newRandKey()
		limit := uint64(10)
		require.NoError(t, fx.SetSpaceLimit(ctx, k, limit))
		// no error
		assert.NoError(t, fx.CheckLimits(ctx, k))

		// add 10 blocks
		bs := testutil.NewRandBlocks(10)
		require.NoError(t, fx.BlocksAdd(ctx, bs))
		fileId := testutil.NewRandCid().String()
		cids, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, k, fileId, cids))
		cids.Release()

		assert.ErrorIs(t, fx.CheckLimits(ctx, k), ErrLimitExceed)
	})
	t.Run("group", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)

		k := newRandKey()
		limit := uint64(10)
		require.NoError(t, fx.SetGroupLimit(ctx, k.GroupId, limit))
		// no error
		assert.NoError(t, fx.CheckLimits(ctx, k))

		// add 10 blocks
		bs := testutil.NewRandBlocks(10)
		require.NoError(t, fx.BlocksAdd(ctx, bs))
		fileId := testutil.NewRandCid().String()
		cids, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, k, fileId, cids))
		cids.Release()

		assert.ErrorIs(t, fx.CheckLimits(ctx, k), ErrLimitExceed)
	})

}

func TestRedisIndex_SetGroupLimit(t *testing.T) {
	t.Run("increase limit", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)

		k := newRandKey()
		limit := uint64(3456)
		require.NoError(t, fx.SetGroupLimit(ctx, k.GroupId, limit))

		info, err := fx.GroupInfo(ctx, k.GroupId)
		require.NoError(t, err)
		assert.Equal(t, limit, info.Limit)
		assert.Equal(t, limit, info.AccountLimit)
	})
	t.Run("decrease limit", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)

		k := newRandKey()
		limit := uint64(100)
		require.NoError(t, fx.SetGroupLimit(ctx, k.GroupId, limit))

		info, err := fx.GroupInfo(ctx, k.GroupId)
		require.NoError(t, err)
		assert.Equal(t, limit, info.Limit)
		assert.Equal(t, limit, info.AccountLimit)
	})
	t.Run("decrease limit with isolated spaces", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)

		k := newRandKey()
		initialLimit := uint64(3000)
		require.NoError(t, fx.SetGroupLimit(ctx, k.GroupId, initialLimit))
		require.NoError(t, fx.SetSpaceLimit(ctx, k, 2000))

		require.NoError(t, fx.SetGroupLimit(ctx, k.GroupId, 1000))

		groupInfo, err := fx.GroupInfo(ctx, k.GroupId)
		require.NoError(t, err)
		assert.Equal(t, uint64(334), groupInfo.Limit)
		assert.Equal(t, uint64(1000), groupInfo.AccountLimit)

		spaceInfo, err := fx.SpaceInfo(ctx, k)
		require.NoError(t, err)
		assert.Equal(t, uint64(666), spaceInfo.Limit)
		t.Log(spaceInfo.Limit)

	})

}

func TestRedisIndex_SetSpaceLimit(t *testing.T) {
	t.Run("isolate empty space", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)

		k := newRandKey()
		groupLimit := uint64(3000)
		spaceLimit := uint64(1000)
		require.NoError(t, fx.SetGroupLimit(ctx, k.GroupId, groupLimit))
		require.NoError(t, fx.SetSpaceLimit(ctx, k, spaceLimit))

		groupInfo, err := fx.GroupInfo(ctx, k.GroupId)
		require.NoError(t, err)
		assert.Equal(t, groupLimit-spaceLimit, groupInfo.Limit)
		assert.Equal(t, groupLimit, groupInfo.AccountLimit)

		spaceInfo, err := fx.SpaceInfo(ctx, k)
		require.NoError(t, err)
		assert.Equal(t, spaceLimit, spaceInfo.Limit)
	})
	t.Run("isolate non-empty space", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)

		k := newRandKey()
		groupLimit := uint64(3000)
		spaceLimit := uint64(1000)
		require.NoError(t, fx.SetGroupLimit(ctx, k.GroupId, groupLimit))

		bs := testutil.NewRandBlocks(3)
		var sumSize uint64
		for _, b := range bs {
			sumSize += uint64(len(b.RawData()))
		}
		k2 := Key{GroupId: k.GroupId, SpaceId: newRandKey().SpaceId}

		require.NoError(t, fx.BlocksAdd(ctx, bs))

		// add first block to space 2
		fileId := testutil.NewRandCid().String()
		cids, err := fx.CidEntriesByBlocks(ctx, bs[:1])
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, k2, fileId, cids))
		cids.Release()

		// add all blocks to space 1
		fileId = testutil.NewRandCid().String()
		cids, err = fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, k, fileId, cids))
		cids.Release()

		// isolate space
		require.NoError(t, fx.SetSpaceLimit(ctx, k, spaceLimit))

		groupInfo, err := fx.GroupInfo(ctx, k.GroupId)
		require.NoError(t, err)
		assert.Equal(t, groupLimit-spaceLimit, groupInfo.Limit)
		assert.Equal(t, groupLimit, groupInfo.AccountLimit)
		// only first common block consume group storage
		assert.Equal(t, uint64(len(bs[0].RawData())), groupInfo.BytesUsage)
		assert.Equal(t, uint64(1), groupInfo.CidsCount)

		spaceInfo, err := fx.SpaceInfo(ctx, k)
		require.NoError(t, err)
		assert.Equal(t, spaceLimit, spaceInfo.Limit)
		assert.Equal(t, sumSize, spaceInfo.BytesUsage)
	})
	t.Run("unite non-empty space", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)

		k := newRandKey()
		groupLimit := uint64(3000)
		spaceLimit := uint64(1000)
		require.NoError(t, fx.SetGroupLimit(ctx, k.GroupId, groupLimit))

		bs := testutil.NewRandBlocks(3)
		var sumSize uint64
		for _, b := range bs {
			sumSize += uint64(len(b.RawData()))
		}
		k2 := Key{GroupId: k.GroupId, SpaceId: newRandKey().SpaceId}

		require.NoError(t, fx.BlocksAdd(ctx, bs))

		// add first block to space 2
		fileId := testutil.NewRandCid().String()
		cids, err := fx.CidEntriesByBlocks(ctx, bs[:1])
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, k2, fileId, cids))
		cids.Release()

		// add all blocks to space 1
		fileId = testutil.NewRandCid().String()
		cids, err = fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, k, fileId, cids))
		cids.Release()

		// isolate space
		require.NoError(t, fx.SetSpaceLimit(ctx, k, spaceLimit))
		// unite space
		require.NoError(t, fx.SetSpaceLimit(ctx, k, 0))

		groupInfo, err := fx.GroupInfo(ctx, k.GroupId)
		require.NoError(t, err)
		assert.Equal(t, groupLimit, groupInfo.Limit)
		assert.Equal(t, groupLimit, groupInfo.AccountLimit)
		// only first common block consume group storage
		assert.Equal(t, sumSize, groupInfo.BytesUsage)
		assert.Equal(t, uint64(3), groupInfo.CidsCount)

		spaceInfo, err := fx.SpaceInfo(ctx, k)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), spaceInfo.Limit)
		assert.Equal(t, sumSize, spaceInfo.BytesUsage)
	})
	t.Run("not enough space", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)

		k := newRandKey()
		groupLimit := uint64(3000)
		require.NoError(t, fx.SetGroupLimit(ctx, k.GroupId, groupLimit))

		bs := testutil.NewRandBlocks(3)
		var sumSize uint64
		for _, b := range bs {
			sumSize += uint64(len(b.RawData()))
		}
		require.NoError(t, fx.BlocksAdd(ctx, bs))

		fileId := testutil.NewRandCid().String()
		cids, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, k, fileId, cids))
		cids.Release()

		assert.EqualError(t, fx.SetSpaceLimit(ctx, k, sumSize-1), fileprotoerr.ErrNotEnoughSpace.Error())
	})
	t.Run("not enough space 2", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)

		k := newRandKey()
		groupLimit := uint64(1000)
		require.NoError(t, fx.SetGroupLimit(ctx, k.GroupId, groupLimit))

		assert.EqualError(t, fx.SetSpaceLimit(ctx, k, groupLimit+1), fileprotoerr.ErrNotEnoughSpace.Error())
	})
}
