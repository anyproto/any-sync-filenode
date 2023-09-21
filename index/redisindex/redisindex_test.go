package redisindex

import (
	"context"
	"github.com/anyproto/any-sync-filenode/index"
	"github.com/anyproto/any-sync-filenode/redisprovider/testredisprovider"
	"github.com/anyproto/any-sync-filenode/testutil"
	"github.com/anyproto/any-sync/app"
	blocks "github.com/ipfs/go-block-format"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strings"
	"testing"
	"time"
)

var ctx = context.Background()

func TestRedisIndex_Exists(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	storeKey := testutil.NewRandSpaceId()
	var bs = make([]blocks.Block, 1)
	for i := range bs {
		bs[i] = testutil.NewRandBlock(rand.Intn(256 * 1024))
	}
	require.NoError(t, fx.Bind(ctx, storeKey, "", bs))
	ex, err := fx.Exists(ctx, bs[0].Cid())
	require.NoError(t, err)
	assert.True(t, ex)
}

func TestRedisIndex_ExistsInSpace(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	storeKey := testutil.NewRandSpaceId()
	var bs = make([]blocks.Block, 2)
	for i := range bs {
		bs[i] = testutil.NewRandBlock(rand.Intn(256 * 1024))
	}
	require.NoError(t, fx.Bind(ctx, storeKey, testutil.NewRandCid().String(), bs[:1]))
	ex, err := fx.ExistsInStorage(ctx, storeKey, testutil.BlocksToKeys(bs))
	require.NoError(t, err)
	assert.Len(t, ex, 1)
}

func TestRedisIndex_IsAllExists(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	storeKey := testutil.NewRandSpaceId()
	fileId := testutil.NewRandCid().String()
	var bs = make([]blocks.Block, 2)
	for i := range bs {
		bs[i] = testutil.NewRandBlock(rand.Intn(256 * 1024))
	}
	require.NoError(t, fx.Bind(ctx, storeKey, fileId, bs[:1]))
	keys := testutil.BlocksToKeys(bs)
	exists, err := fx.IsAllExists(ctx, keys)
	require.NoError(t, err)
	assert.False(t, exists)
	exists, err = fx.IsAllExists(ctx, keys[:1])
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRedisIndex_GetNonExistentBlocks(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	storeKey := testutil.NewRandSpaceId()
	fileId := testutil.NewRandCid().String()
	var bs = make([]blocks.Block, 2)
	for i := range bs {
		bs[i] = testutil.NewRandBlock(rand.Intn(256 * 1024))
	}
	require.NoError(t, fx.Bind(ctx, storeKey, fileId, bs[:1]))

	nonExistent, err := fx.GetNonExistentBlocks(ctx, bs)
	require.NoError(t, err)
	require.Len(t, nonExistent, 1)
	assert.Equal(t, bs[1:], nonExistent)
}

func TestRedisIndex_SpaceSize(t *testing.T) {
	t.Run("space not found", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		storeKey := testutil.NewRandSpaceId()
		size, err := fx.StorageSize(ctx, storeKey)
		require.NoError(t, err)
		assert.Empty(t, size)
	})
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		var storeKey = testutil.NewRandSpaceId()
		var expectedSize int
		fileId := testutil.NewRandCid().String()
		var bs = make([]blocks.Block, 2)
		for i := range bs {
			bs[i] = testutil.NewRandBlock(1024)
			expectedSize += 1024
		}
		require.NoError(t, fx.AddBlocks(ctx, bs))
		require.NoError(t, fx.Bind(ctx, storeKey, fileId, bs))

		size, err := fx.StorageSize(ctx, storeKey)
		require.NoError(t, err)
		assert.Equal(t, expectedSize, int(size))
	})
}

func TestRedisIndex_Lock(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	var bs = make([]blocks.Block, 3)
	for i := range bs {
		bs[i] = testutil.NewRandBlock(rand.Intn(1024))
	}

	unlock, err := fx.Lock(ctx, testutil.BlocksToKeys(bs[1:]))
	require.NoError(t, err)
	tCtx, cancel := context.WithTimeout(ctx, time.Second/2)
	defer cancel()
	_, err = fx.Lock(tCtx, testutil.BlocksToKeys(bs))
	require.Error(t, err)
	unlock()
}

func TestRedisIndex_MoveStorage(t *testing.T) {
	const (
		oldKey = "oldKey"
		newKey = "newKey"
	)
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		require.NoError(t, fx.Bind(ctx, oldKey, "fid", testutil.NewRandBlocks(1)))
		require.NoError(t, fx.MoveStorage(ctx, oldKey, newKey))
	})
	t.Run("err storage not found", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		assert.EqualError(t, fx.MoveStorage(ctx, oldKey, newKey), index.ErrStorageNotFound.Error())
	})
	t.Run("err taget exists", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		require.NoError(t, fx.Bind(ctx, oldKey, "fid", testutil.NewRandBlocks(1)))
		require.NoError(t, fx.Bind(ctx, newKey, "fid", testutil.NewRandBlocks(1)))
		assert.EqualError(t, fx.MoveStorage(ctx, oldKey, newKey), index.ErrTargetStorageExists.Error())
	})
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
		storeKey := testutil.NewRandSpaceId()
		fileId := testutil.NewRandCid().String()
		require.NoError(t, fx.Bind(ctx, storeKey, fileId, bs))
		t.Logf("bound %d cid for a %v", len(bs), time.Since(st))
		st = time.Now()
		sz, err := fx.StorageSize(ctx, storeKey)
		require.NoError(t, err)
		t.Logf("space size is %d, dur: %v", sz, time.Since(st))
	}
	info, err := fx.cl.Info(ctx, "memory").Result()
	require.NoError(t, err)
	infoS := strings.Split(info, "\n")
	for _, i := range infoS {
		if strings.HasPrefix(i, "used_memory_human") {
			t.Log(i)
		}
	}
}

func TestRedisIndex_AddBlocks(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	var bs = make([]blocks.Block, 3)
	for i := range bs {
		bs[i] = testutil.NewRandBlock(rand.Intn(1024))
	}

	require.NoError(t, fx.AddBlocks(ctx, bs))

	for _, b := range bs {
		ex, err := fx.Exists(ctx, b.Cid())
		require.NoError(t, err)
		assert.True(t, ex)
	}
}

func newFixture(t require.TestingT) (fx *fixture) {
	fx = &fixture{
		redisIndex: New().(*redisIndex),
		a:          new(app.App),
	}
	fx.a.Register(testredisprovider.NewTestRedisProvider()).Register(fx.redisIndex)
	require.NoError(t, fx.a.Start(ctx))
	return
}

type fixture struct {
	*redisIndex
	a *app.App
}

func (fx *fixture) Finish(t require.TestingT) {
	require.NoError(t, fx.a.Close(ctx))
}
