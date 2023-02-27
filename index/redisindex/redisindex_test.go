package redisindex

import (
	"context"
	"github.com/anytypeio/any-sync-filenode/index"
	"github.com/anytypeio/any-sync-filenode/redisprovider/testredisprovider"
	"github.com/anytypeio/any-sync-filenode/testutil"
	"github.com/anytypeio/any-sync/app"
	"github.com/ipfs/go-libipfs/blocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

var ctx = context.Background()

func TestRedisIndex_Bind(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	spaceId1 := testutil.NewRandSpaceId()
	spaceId2 := testutil.NewRandSpaceId()
	var bs = make([]blocks.Block, 10)
	var size int
	for i := range bs {
		bs[i] = testutil.NewRandBlock(rand.Intn(256 * 1024))
		size += len(bs[i].RawData())
	}

	require.NoError(t, fx.Bind(ctx, spaceId1, bs))

	keys := testutil.BlocksToKeys(bs)
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
	assert.Equal(t, testutil.BlocksToKeys(bs[:5]), exKeys2)
	ss, err := fx.SpaceSize(ctx, spaceId1)
	require.NoError(t, err)
	assert.Equal(t, uint64(size), ss)
}

func TestRedisIndex_UnBind(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	spaceId1 := testutil.NewRandSpaceId()
	spaceId2 := testutil.NewRandSpaceId()
	var bs = make([]blocks.Block, 10)
	for i := range bs {
		bs[i] = testutil.NewRandBlock(rand.Intn(256 * 1024))
	}

	require.NoError(t, fx.Bind(ctx, spaceId1, bs))
	require.NoError(t, fx.Bind(ctx, spaceId2, bs[:4]))

	keys := testutil.BlocksToKeys(bs)
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
	ss, err := fx.SpaceSize(ctx, spaceId1)
	require.NoError(t, err)
	assert.Empty(t, ss)
}

func TestRedisIndex_Exists(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	spaceId1 := testutil.NewRandSpaceId()
	var bs = make([]blocks.Block, 1)
	for i := range bs {
		bs[i] = testutil.NewRandBlock(rand.Intn(256 * 1024))
	}
	require.NoError(t, fx.Bind(ctx, spaceId1, bs))
	ex, err := fx.Exists(ctx, bs[0].Cid())
	require.NoError(t, err)
	assert.True(t, ex)
}

func TestRedisIndex_FilterExistingOnly(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	spaceId1 := testutil.NewRandSpaceId()
	var bs = make([]blocks.Block, 2)
	for i := range bs {
		bs[i] = testutil.NewRandBlock(rand.Intn(256 * 1024))
	}
	require.NoError(t, fx.Bind(ctx, spaceId1, bs[:1]))
	keys := testutil.BlocksToKeys(bs)

	exists, err := fx.FilterExistingOnly(ctx, keys)
	require.NoError(t, err)
	assert.Len(t, exists, 1)
}

func TestRedisIndex_GetNonExistentBlocks(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	spaceId1 := testutil.NewRandSpaceId()
	var bs = make([]blocks.Block, 2)
	for i := range bs {
		bs[i] = testutil.NewRandBlock(rand.Intn(256 * 1024))
	}
	require.NoError(t, fx.Bind(ctx, spaceId1, bs[:1]))

	nonExistent, err := fx.GetNonExistentBlocks(ctx, bs)
	require.NoError(t, err)
	require.Len(t, nonExistent, 1)
	assert.Equal(t, bs[1:], nonExistent)
}

func TestRedisIndex_Fuzzy(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	spaceId1 := testutil.NewRandSpaceId()
	var bs = make([]blocks.Block, 100)
	for i := range bs {
		bs[i] = testutil.NewRandBlock(rand.Intn(1024))
	}

	var stopCh = make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	// bind goroutine
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopCh:
				return
			default:
			}
			var bs2 = make([]blocks.Block, len(bs))
			copy(bs2, bs)
			rand.Shuffle(len(bs2), func(i, j int) {
				bs2[i], bs2[j] = bs2[j], bs2[i]
			})
			require.NoError(t, fx.Bind(ctx, spaceId1, bs2[:rand.Intn(20)+1]))
		}
	}()

	// unbind goroutine
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopCh:
				return
			default:
			}
			var bs2 = make([]blocks.Block, len(bs))
			copy(bs2, bs)
			rand.Shuffle(len(bs2), func(i, j int) {
				bs2[i], bs2[j] = bs2[j], bs2[i]
			})
			_, err := fx.UnBind(ctx, spaceId1, testutil.BlocksToKeys(bs2[:rand.Intn(20)+1]))
			require.NoError(t, err)
		}
	}()

	time.Sleep(time.Second)
	close(stopCh)
	wg.Wait()
	time.Sleep(time.Second / 100)
	_, err := fx.UnBind(ctx, spaceId1, testutil.BlocksToKeys(bs))
	require.NoError(t, err)
	size, err := fx.SpaceSize(ctx, spaceId1)
	require.NoError(t, err)
	assert.Empty(t, size)
}

func Test100KCids(t *testing.T) {
	t.Skip()
	fx := newFixture(t)
	defer fx.Finish(t)
	for i := 0; i < 10; i++ {
		st := time.Now()
		var bs = make([]blocks.Block, 10000)
		for n := range bs {
			bs[n] = testutil.NewRandBlock(rand.Intn(256))
		}
		spaceId := testutil.NewRandSpaceId()
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
