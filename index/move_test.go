package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anyproto/any-sync-filenode/testutil"
)

func TestRedisIndex_CheckOwnership(t *testing.T) {
	t.Run("same owner", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		err := fx.CheckOwnership(ctx, Key{"o1", "s1"}, "s1", 0)
		require.NoError(t, err)
		err = fx.CheckOwnership(ctx, Key{"o1", "s1"}, "", 0)
		require.NoError(t, err)
		err = fx.CheckOwnership(ctx, Key{"o1", "s1"}, "", 22)
		require.NoError(t, err)
	})
	t.Run("change owner", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		err := fx.CheckOwnership(ctx, Key{"alice", "s1"}, "sn", 0)
		require.NoError(t, err)
		err = fx.CheckOwnership(ctx, Key{GroupId: "bob", SpaceId: "s1"}, "sn", 1)
		require.NoError(t, err)
	})
	t.Run("old index", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		err := fx.CheckOwnership(ctx, Key{"alice", "s1"}, "sn", 11)
		require.NoError(t, err)
		err = fx.CheckOwnership(ctx, Key{"bob", "s1"}, "sn", 9)
		require.NoError(t, err)
	})
	t.Run("no prev owner", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		err := fx.CheckOwnership(ctx, Key{"alice", "s1"}, "", 12)
		assert.Error(t, err)
	})
}

func TestRedisIndex_Move(t *testing.T) {
	var test = func(aliceKey, bobKey Key) {
		fx := newFixture(t)
		defer fx.Finish(t)

		aliceBS := testutil.NewRandBlocks(3)
		require.NoError(t, fx.BlocksAdd(ctx, aliceBS))
		aliceCids, err := fx.CidEntriesByBlocks(ctx, aliceBS)
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, aliceKey, "f1", aliceCids))
		aliceCids.Release()

		bobBS := testutil.NewRandBlocks(3)
		require.NoError(t, fx.BlocksAdd(ctx, bobBS))
		bobCids, err := fx.CidEntriesByBlocks(ctx, bobBS)
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, bobKey, "f1", bobCids))
		bobCids.Release()

		aliceGroupBefore, err := fx.GroupInfo(ctx, aliceKey.GroupId)
		require.NoError(t, err)
		bobGroupBefore, err := fx.GroupInfo(ctx, bobKey.GroupId)
		require.NoError(t, err)
		spaceBefore, err := fx.SpaceInfo(ctx, bobKey)
		require.NoError(t, err)

		require.NoError(t, fx.Move(ctx, Key{aliceKey.GroupId, bobKey.SpaceId}, bobKey))

		aliceGroupAfter, err := fx.GroupInfo(ctx, aliceKey.GroupId)
		require.NoError(t, err)
		bobGroupAfter, err := fx.GroupInfo(ctx, bobKey.GroupId)
		require.NoError(t, err)
		spaceAfter, err := fx.SpaceInfo(ctx, Key{aliceKey.GroupId, bobKey.SpaceId})
		require.NoError(t, err)

		assert.Equal(t, int(aliceGroupBefore.BytesUsage+spaceBefore.BytesUsage), int(aliceGroupAfter.BytesUsage))
		assert.Greater(t, int(bobGroupBefore.BytesUsage), 0)
		assert.NotEqual(t, bobGroupBefore.BytesUsage, bobGroupAfter.BytesUsage)
		assert.Equal(t, int(bobGroupAfter.BytesUsage), 0)
		assert.Equal(t, int(bobGroupAfter.CidsCount), 0)
		assert.Len(t, bobGroupAfter.SpaceIds, 0)
		assert.Len(t, aliceGroupAfter.SpaceIds, 2)
		assert.Equal(t, int(spaceAfter.BytesUsage), int(spaceBefore.BytesUsage))
	}

	t.Run("move", func(t *testing.T) {
		var srcKey = Key{"g1", "s1"}
		var dstKey = Key{"g2", "s2"}
		test(srcKey, dstKey)
	})
	t.Run("move same hash", func(t *testing.T) {
		// these groups has an identical xxhash
		var srcKey = Key{"3325572473", "s1"}
		var dstKey = Key{"8117876798", "s2"}
		test(srcKey, dstKey)
	})
}
