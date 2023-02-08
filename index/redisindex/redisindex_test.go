package redisindex

import (
	"context"
	"github.com/anytypeio/any-sync-filenode/index"
	"github.com/anytypeio/any-sync-filenode/redisprovider/testredisprovider"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/util/cidutil"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-libipfs/blocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"math/rand"
	"strings"
	"testing"
	"time"
)

var ctx = context.Background()

func TestRedisIndex_Bind(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	spaceId1 := newRandSpaceId()
	spaceId2 := newRandSpaceId()
	var bs = make([]blocks.Block, 10)
	for i := range bs {
		bs[i] = newRandBlock(rand.Intn(256 * 1024))
	}

	require.NoError(t, fx.Bind(ctx, spaceId1, bs))

	keys := blocksToCids(bs)
	exKeys, err := fx.ExistsInSpace(ctx, spaceId1, keys)
	require.NoError(t, err)
	assert.Equal(t, keys, exKeys)

	exKeys2, err := fx.ExistsInSpace(ctx, spaceId2, keys)
	require.NoError(t, err)
	assert.Empty(t, exKeys2)

	require.NoError(t, fx.Bind(ctx, spaceId1, bs))
	require.NoError(t, fx.Bind(ctx, spaceId2, bs[:5]))
	exKeys2, err = fx.ExistsInSpace(ctx, spaceId2, keys)
	require.NoError(t, err)
	assert.Equal(t, blocksToCids(bs[:5]), exKeys2)
}

func TestRedisIndex_UnBind(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	spaceId1 := newRandSpaceId()
	spaceId2 := newRandSpaceId()
	var bs = make([]blocks.Block, 10)
	for i := range bs {
		bs[i] = newRandBlock(rand.Intn(256 * 1024))
	}

	require.NoError(t, fx.Bind(ctx, spaceId1, bs))
	require.NoError(t, fx.Bind(ctx, spaceId2, bs[:4]))

	keys := blocksToCids(bs)
	toDelete, err := fx.UnBind(ctx, spaceId1, keys)
	require.NoError(t, err)
	assert.Len(t, toDelete, 6)
	for _, deleted := range toDelete {
		ex, e := fx.Exists(ctx, deleted)
		require.NoError(t, e)
		assert.False(t, ex)
	}

	toDelete, err = fx.UnBind(ctx, spaceId2, keys)
	require.NoError(t, err)
	assert.Len(t, toDelete, 4)
	for _, deleted := range toDelete {
		ex, e := fx.Exists(ctx, deleted)
		require.NoError(t, e)
		assert.False(t, ex)
	}
}

func TestRedisIndex_Exists(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	spaceId1 := newRandSpaceId()
	var bs = make([]blocks.Block, 1)
	for i := range bs {
		bs[i] = newRandBlock(rand.Intn(256 * 1024))
	}
	require.NoError(t, fx.Bind(ctx, spaceId1, bs))
	ex, err := fx.Exists(ctx, bs[0].Cid())
	require.NoError(t, err)
	assert.True(t, ex)
}

func TestRedisIndex_FilterExistingOnly(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	spaceId1 := newRandSpaceId()
	var bs = make([]blocks.Block, 2)
	for i := range bs {
		bs[i] = newRandBlock(rand.Intn(256 * 1024))
	}
	require.NoError(t, fx.Bind(ctx, spaceId1, bs[:1]))
	keys := blocksToCids(bs)

	exists, err := fx.FilterExistingOnly(ctx, keys)
	require.NoError(t, err)
	assert.Len(t, exists, 1)
}

func TestRedisIndex_GetNonExistentBlocks(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	spaceId1 := newRandSpaceId()
	var bs = make([]blocks.Block, 2)
	for i := range bs {
		bs[i] = newRandBlock(rand.Intn(256 * 1024))
	}
	require.NoError(t, fx.Bind(ctx, spaceId1, bs[:1]))

	nonExistent, err := fx.GetNonExistentBlocks(ctx, bs)
	require.NoError(t, err)
	require.Len(t, nonExistent, 1)
	assert.Equal(t, bs[1:], nonExistent)
}

func Test100KCids(t *testing.T) {
	t.Skip()
	fx := newFixture(t)
	defer fx.Finish(t)
	for i := 0; i < 10; i++ {
		st := time.Now()
		var bs = make([]blocks.Block, 10000)
		for n := range bs {
			bs[n] = newRandBlock(rand.Intn(256))
		}
		spaceId := newRandSpaceId()
		require.NoError(t, fx.Bind(ctx, spaceId, bs))
		t.Logf("bound %d cid for a %v", len(bs), time.Since(st))
		st = time.Now()
		sz, err := fx.SpaceSize(ctx, spaceId)
		require.NoError(t, err)
		t.Logf("space size is %d, dur: %v", sz, time.Since(st))
	}
	info, err := fx.Index.(*redisIndex).cl.Info(ctx, "memory").Result()
	require.NoError(t, err)
	infoS := strings.Split(info, "\n")
	for _, i := range infoS {
		if strings.HasPrefix(i, "used_memory_human") {
			t.Log(i)
		}
	}
}

func newFixture(t require.TestingT) (fx *fixture) {
	fx = &fixture{
		Index: New(),
		a:     new(app.App),
	}
	fx.a.Register(testredisprovider.NewTestRedisProvider()).Register(fx.Index)
	require.NoError(t, fx.a.Start(ctx))
	return
}

type fixture struct {
	index.Index
	a *app.App
}

func (fx *fixture) Finish(t require.TestingT) {
	require.NoError(t, fx.a.Close(ctx))
}

func newRandSpaceId() string {
	b := newRandBlock(256)
	return b.Cid().String() + ".123456"
}

func newRandBlock(size int) blocks.Block {
	var p = make([]byte, size)
	_, err := io.ReadFull(rand.New(rand.NewSource(time.Now().UnixNano())), p)
	if err != nil {
		panic("can't fill testdata from rand")
	}
	c, _ := cidutil.NewCidFromBytes(p)
	b, _ := blocks.NewBlockWithCid(p, cid.MustParse(c))
	return b
}

func blocksToCids(bs []blocks.Block) (cids []cid.Cid) {
	cids = make([]cid.Cid, len(bs))
	for i, b := range bs {
		cids[i] = b.Cid()
	}
	return
}
