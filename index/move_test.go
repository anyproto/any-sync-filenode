package index

import (
	"strings"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anyproto/any-sync-filenode/testutil"
)

func TestRedisIndex_CheckOwnership(t *testing.T) {
	t.Run("same owner", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		err := fx.CheckAndMoveOwnership(ctx, Key{"o1", "s1"}, "s1", 0)
		require.NoError(t, err)
		err = fx.CheckAndMoveOwnership(ctx, Key{"o1", "s1"}, "", 0)
		require.NoError(t, err)
		err = fx.CheckAndMoveOwnership(ctx, Key{"o1", "s1"}, "", 22)
		require.NoError(t, err)
	})
	t.Run("change owner", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		err := fx.CheckAndMoveOwnership(ctx, Key{"alice", "s1"}, "sn", 0)
		require.NoError(t, err)
		err = fx.CheckAndMoveOwnership(ctx, Key{GroupId: "bob", SpaceId: "s1"}, "sn", 1)
		require.NoError(t, err)
	})
	t.Run("old index", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		err := fx.CheckAndMoveOwnership(ctx, Key{"alice", "s1"}, "sn", 11)
		require.NoError(t, err)
		err = fx.CheckAndMoveOwnership(ctx, Key{"bob", "s1"}, "sn", 9)
		require.NoError(t, err)
	})
	t.Run("no prev owner", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		err := fx.CheckAndMoveOwnership(ctx, Key{"alice", "s1"}, "", 12)
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

		// file entries and per-cid refcounts must survive the move
		movedKey := Key{aliceKey.GroupId, bobKey.SpaceId}
		fileIds, err := fx.FilesList(ctx, movedKey)
		require.NoError(t, err)
		assert.Equal(t, []string{"f1"}, fileIds)

		fileInfos, err := fx.FileInfo(ctx, movedKey, "f1")
		require.NoError(t, err)
		require.Len(t, fileInfos, 1)
		assert.Equal(t, int(spaceBefore.BytesUsage), int(fileInfos[0].BytesUsage))
		assert.Equal(t, len(bobBS), int(fileInfos[0].CidsCount))

		bobCidList := make([]cid.Cid, len(bobBS))
		for i, b := range bobBS {
			bobCidList[i] = b.Cid()
		}
		inSpace, err := fx.CidExistsInSpace(ctx, movedKey, bobCidList)
		require.NoError(t, err)
		assert.Len(t, inSpace, len(bobBS))

		// unbind after the move must release the usage from the new group
		require.NoError(t, fx.FileUnbind(ctx, movedKey, "f1"))
		aliceGroupFinal, err := fx.GroupInfo(ctx, aliceKey.GroupId)
		require.NoError(t, err)
		assert.Equal(t, int(aliceGroupBefore.BytesUsage), int(aliceGroupFinal.BytesUsage))
		spaceFinal, err := fx.SpaceInfo(ctx, movedKey)
		require.NoError(t, err)
		assert.Equal(t, 0, int(spaceFinal.BytesUsage))
		assert.Equal(t, 0, int(spaceFinal.FileCount))
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

	t.Run("shared cid keeps file-binding refcounts", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		var (
			srcKey   = Key{"g1", "s1"}
			movedKey = Key{"g2", "s1"}
		)

		bs := testutil.NewRandBlocks(3)
		require.NoError(t, fx.BlocksAdd(ctx, bs))
		cids, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, srcKey, "f1", cids))
		cids.Release()
		// f2 shares bs[0] with f1
		sharedCids, err := fx.CidEntriesByBlocks(ctx, bs[:1])
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, srcKey, "f2", sharedCids))
		sharedCids.Release()

		require.NoError(t, fx.Move(ctx, movedKey, srcKey))

		// group refcount must equal the number of file bindings, not 1 per distinct cid
		val, err := fx.cl.HGet(ctx, GroupKey(movedKey), CidKey(bs[0].Cid())).Result()
		require.NoError(t, err)
		assert.Equal(t, "2", val)

		// unbinding both files must fully release the group usage
		require.NoError(t, fx.FileUnbind(ctx, movedKey, "f1", "f2"))
		groupAfter, err := fx.GroupInfo(ctx, movedKey.GroupId)
		require.NoError(t, err)
		assert.Equal(t, 0, int(groupAfter.BytesUsage))
		fields, err := fx.cl.HKeys(ctx, GroupKey(movedKey)).Result()
		require.NoError(t, err)
		for _, f := range fields {
			assert.False(t, strings.HasPrefix(f, "c:"), "leftover group cid ref %s", f)
		}
	})

	t.Run("repeated move is idempotent", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		var (
			srcKey   = Key{"g1", "s1"}
			movedKey = Key{"g2", "s1"}
		)

		bs := testutil.NewRandBlocks(3)
		require.NoError(t, fx.BlocksAdd(ctx, bs))
		cids, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, srcKey, "f1", cids))
		cids.Release()

		spaceBefore, err := fx.SpaceInfo(ctx, srcKey)
		require.NoError(t, err)

		require.NoError(t, fx.Move(ctx, movedKey, srcKey))
		// simulate a crash before the owner record was persisted: Move runs again
		require.NoError(t, fx.Move(ctx, movedKey, srcKey))

		spaceAfter, err := fx.SpaceInfo(ctx, movedKey)
		require.NoError(t, err)
		assert.Equal(t, int(spaceBefore.BytesUsage), int(spaceAfter.BytesUsage))
		assert.Equal(t, int(spaceBefore.FileCount), int(spaceAfter.FileCount))

		val, err := fx.cl.HGet(ctx, GroupKey(movedKey), CidKey(bs[0].Cid())).Result()
		require.NoError(t, err)
		assert.Equal(t, "1", val)

		fileIds, err := fx.FilesList(ctx, movedKey)
		require.NoError(t, err)
		assert.Equal(t, []string{"f1"}, fileIds)
	})

	t.Run("interrupted move is finished by retry", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		var (
			srcKey   = Key{"g1", "s1"}
			movedKey = Key{"g2", "s1"}
		)

		bs := testutil.NewRandBlocks(3)
		require.NoError(t, fx.BlocksAdd(ctx, bs))
		cids, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, srcKey, "f1", cids))
		cids.Release()
		// f2 shares bs[0] with f1
		sharedCids, err := fx.CidEntriesByBlocks(ctx, bs[:1])
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, srcKey, "f2", sharedCids))
		sharedCids.Release()

		spaceBefore, err := fx.SpaceInfo(ctx, srcKey)
		require.NoError(t, err)

		// apply only the destination half of the move, simulating a crash
		// before the src cleanup
		srcEntry, srcRelease, err := fx.AcquireSpace(ctx, srcKey)
		require.NoError(t, err)
		destGroup, err := fx.getGroupEntry(ctx, movedKey)
		require.NoError(t, err)
		destSpaceExists, destSRelease, err := fx.AcquireKey(ctx, SpaceKey(movedKey))
		require.NoError(t, err)
		plan, err := fx.prepareMove(ctx, movedKey, srcKey, srcEntry, destGroup, destSpaceExists)
		require.NoError(t, err)
		require.NoError(t, fx.applyMoveDest(ctx, plan))
		plan.cids.Release()
		destSRelease()
		srcRelease()

		// both keys exist now; the retry must complete the move without
		// applying the refcounts twice
		require.NoError(t, fx.Move(ctx, movedKey, srcKey))

		val, err := fx.cl.HGet(ctx, GroupKey(movedKey), CidKey(bs[0].Cid())).Result()
		require.NoError(t, err)
		assert.Equal(t, "2", val)

		srcExists, err := fx.cl.Exists(ctx, SpaceKey(srcKey)).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), srcExists)

		spaceAfter, err := fx.SpaceInfo(ctx, movedKey)
		require.NoError(t, err)
		assert.Equal(t, int(spaceBefore.BytesUsage), int(spaceAfter.BytesUsage))

		destGroupAfter, err := fx.GroupInfo(ctx, movedKey.GroupId)
		require.NoError(t, err)
		assert.Equal(t, int(spaceBefore.BytesUsage), int(destGroupAfter.BytesUsage))
		srcGroupAfter, err := fx.GroupInfo(ctx, srcKey.GroupId)
		require.NoError(t, err)
		assert.Equal(t, 0, int(srcGroupAfter.BytesUsage))
		assert.Len(t, srcGroupAfter.SpaceIds, 0)

		// unbinding everything after the recovered move releases all usage
		require.NoError(t, fx.FileUnbind(ctx, movedKey, "f1", "f2"))
		destGroupFinal, err := fx.GroupInfo(ctx, movedKey.GroupId)
		require.NoError(t, err)
		assert.Equal(t, 0, int(destGroupFinal.BytesUsage))
		fields, err := fx.cl.HKeys(ctx, GroupKey(movedKey)).Result()
		require.NoError(t, err)
		for _, f := range fields {
			assert.False(t, strings.HasPrefix(f, "c:"), "leftover group cid ref %s", f)
		}
	})

	t.Run("isolated space move leaves groups untouched", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		var (
			srcKey   = Key{"g1", "s1"}
			movedKey = Key{"g2", "s1"}
		)

		// isolate the space before any binds: binds don't count against the group
		se, err := fx.getSpaceEntry(ctx, srcKey)
		require.NoError(t, err)
		se.Limit = 2048
		_, err = fx.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			se.Save(ctx, srcKey, pipe)
			return nil
		})
		require.NoError(t, err)

		bs := testutil.NewRandBlocks(3)
		require.NoError(t, fx.BlocksAdd(ctx, bs))
		cids, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, srcKey, "f1", cids))
		cids.Release()

		spaceBefore, err := fx.SpaceInfo(ctx, srcKey)
		require.NoError(t, err)
		require.Greater(t, int(spaceBefore.BytesUsage), 0)

		require.NoError(t, fx.Move(ctx, movedKey, srcKey))

		// neither group may hold cid refs or usage for an isolated space
		for _, groupId := range []string{srcKey.GroupId, movedKey.GroupId} {
			g, err := fx.GroupInfo(ctx, groupId)
			require.NoError(t, err)
			assert.Equal(t, 0, int(g.BytesUsage), "group %s", groupId)
			fields, err := fx.cl.HKeys(ctx, GroupKey(Key{GroupId: groupId})).Result()
			require.NoError(t, err)
			for _, f := range fields {
				assert.False(t, strings.HasPrefix(f, "c:"), "group %s leftover cid ref %s", groupId, f)
			}
		}

		// space data survives the move intact
		spaceAfter, err := fx.SpaceInfo(ctx, movedKey)
		require.NoError(t, err)
		assert.Equal(t, int(spaceBefore.BytesUsage), int(spaceAfter.BytesUsage))
		fileIds, err := fx.FilesList(ctx, movedKey)
		require.NoError(t, err)
		assert.Equal(t, []string{"f1"}, fileIds)
	})
}
