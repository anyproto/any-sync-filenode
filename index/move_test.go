package index

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anyproto/any-sync-filenode/testutil"
)

func TestRedisIndex_Move(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	var (
		oldKey = newRandKey()
		newKey = Key{
			GroupId: oldKey.GroupId + "_new",
			SpaceId: oldKey.SpaceId,
		}
	)

	for i := 0; i < 10; i++ {
		bs := testutil.NewRandBlocks(10)
		var sumSize uint64
		for _, b := range bs {
			sumSize += uint64(len(b.RawData()))
		}
		fileId := testutil.NewRandCid().String()
		require.NoError(t, fx.BlocksAdd(ctx, bs))
		cids, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, oldKey, fileId, cids))
		cids.Release()
	}

	oldGroupInfo, err := fx.GroupInfo(ctx, oldKey.GroupId)
	require.NoError(t, err)
	oldSpaceInfo, err := fx.SpaceInfo(ctx, oldKey)
	require.NoError(t, err)

	st := time.Now()
	require.NoError(t, fx.Move(ctx, oldKey, newKey))
	t.Log(time.Since(st))

	newGroupInfo, err := fx.GroupInfo(ctx, newKey.GroupId)
	require.NoError(t, err)
	newSpaceInfo, err := fx.SpaceInfo(ctx, newKey)
	require.NoError(t, err)

	assert.Equal(t, oldGroupInfo, newGroupInfo)
	assert.Equal(t, oldSpaceInfo, newSpaceInfo)

	oldGroupInfo, err = fx.GroupInfo(ctx, oldKey.GroupId)
	require.NoError(t, err)
	oldSpaceInfo, err = fx.SpaceInfo(ctx, oldKey)
	require.NoError(t, err)

	assert.Empty(t, oldGroupInfo)
	assert.Empty(t, oldSpaceInfo)

}
