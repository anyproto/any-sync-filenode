package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anyproto/any-sync-filenode/testutil"
)

func TestRedisIndex_Check(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	// add some data
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
	require.NoError(t, fx.FileBind(ctx, key, fileId2, cidsB))
	cidsB.Release()

	t.Run("fix space+group size", func(t *testing.T) {
		se, err := fx.getSpaceEntry(ctx, key)
		require.NoError(t, err)
		se.Size_ += 100
		seData, _ := se.Marshal()
		require.NoError(t, fx.cl.HSet(ctx, SpaceKey(key), infoKey, seData).Err())

		ge, err := fx.getGroupEntry(ctx, key)
		require.NoError(t, err)
		ge.Size_ += 100
		geData, _ := ge.Marshal()
		require.NoError(t, fx.cl.HSet(ctx, GroupKey(key), infoKey, geData).Err())

		fixRes, err := fx.Check(ctx, key, true)
		require.NoError(t, err)
		assert.Len(t, fixRes, 2)
		fixRes, err = fx.Check(ctx, key, false)
		require.NoError(t, err)
		assert.Len(t, fixRes, 0)
	})
	t.Run("fix space+group cids", func(t *testing.T) {
		require.NoError(t, fx.cl.HSet(ctx, SpaceKey(key), "c:"+bs[1].Cid().String(), 33).Err())
		require.NoError(t, fx.cl.HSet(ctx, SpaceKey(key), "c:"+testutil.NewRandCid().String(), 33).Err())
		require.NoError(t, fx.cl.HSet(ctx, GroupKey(key), "c:"+bs[1].Cid().String(), 43).Err())
		require.NoError(t, fx.cl.HSet(ctx, GroupKey(key), "c:"+testutil.NewRandCid().String(), 43).Err())

		fixRes, err := fx.Check(ctx, key, true)
		require.NoError(t, err)
		assert.Len(t, fixRes, 4)
		fixRes, err = fx.Check(ctx, key, false)
		require.NoError(t, err)
		assert.Len(t, fixRes, 0)
	})

	t.Run("fix cid", func(t *testing.T) {
		ce, err := fx.getCidEntry(ctx, bs[0].Cid())
		require.NoError(t, err)
		ce.Refs = 0
		require.NoError(t, ce.Save(ctx, fx.cl))

		fixRes, err := fx.Check(ctx, key, true)
		require.NoError(t, err)
		assert.Len(t, fixRes, 1)
		for _, f := range fixRes {
			t.Log(f.Description)
		}
		fixRes, err = fx.Check(ctx, key, false)
		require.NoError(t, err)
		for _, f := range fixRes {
			t.Log(f.Description)
		}
		assert.Len(t, fixRes, 0)
	})

}
