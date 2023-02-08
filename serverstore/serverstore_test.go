package serverstore

import (
	"context"
	"github.com/anytypeio/any-sync-filenode/index/mock_index"
	"github.com/anytypeio/any-sync-filenode/index/redisindex"
	"github.com/anytypeio/any-sync-filenode/serverstore/mock_serverstore"
	"github.com/anytypeio/any-sync-filenode/testutil"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/commonfile/fileblockstore"
	"github.com/golang/mock/gomock"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-libipfs/blocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var ctx = context.Background()

func TestServerStore_Add(t *testing.T) {
	t.Run("no spaceId", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		require.Error(t, fx.Add(ctx, []blocks.Block{testutil.NewRandBlock(10)}))
	})
	t.Run("all blocks exists", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		var spaceId = "space1"
		sCtx := fileblockstore.CtxWithSpaceId(ctx, spaceId)
		var bs = make([]blocks.Block, 10)
		for i := range bs {
			bs[i] = testutil.NewRandBlock(10)
		}
		fx.index.EXPECT().GetNonExistentBlocks(sCtx, bs).Return(nil, nil)
		fx.index.EXPECT().Bind(sCtx, spaceId, bs)
		require.NoError(t, fx.Add(sCtx, bs))
	})
	t.Run("partial upload", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		var spaceId = "space1"
		sCtx := fileblockstore.CtxWithSpaceId(ctx, spaceId)
		var bs = make([]blocks.Block, 10)
		for i := range bs {
			bs[i] = testutil.NewRandBlock(10)
		}
		fx.index.EXPECT().GetNonExistentBlocks(sCtx, bs).Return(bs[:5], nil)
		fx.store.EXPECT().Add(sCtx, bs[:5])
		fx.index.EXPECT().Bind(sCtx, spaceId, bs)
		require.NoError(t, fx.Add(sCtx, bs))
	})
}

func TestServerStore_Get(t *testing.T) {
	t.Run("not exists", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		b := testutil.NewRandBlock(1)
		fx.index.EXPECT().Exists(ctx, b.Cid()).Return(false, nil)
		_, err := fx.Get(ctx, b.Cid())
		require.Error(t, err)
	})
	t.Run("exists", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		b := testutil.NewRandBlock(1)
		fx.index.EXPECT().Exists(ctx, b.Cid()).Return(true, nil)
		fx.store.EXPECT().Get(ctx, b.Cid()).Return(b, nil)
		resB, err := fx.Get(ctx, b.Cid())
		require.NoError(t, err)
		assert.Equal(t, b, resB)
	})
}

func TestServerStore_GetMany(t *testing.T) {
	t.Run("not exists", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		bs := make([]blocks.Block, 2)
		for i := range bs {
			bs[i] = testutil.NewRandBlock(10)
		}
		fx.index.EXPECT().FilterExistingOnly(ctx, testutil.BlocksToKeys(bs)).Return(nil, nil)
		res := fx.GetMany(ctx, testutil.BlocksToKeys(bs))
		assert.EqualValues(t, closedResult, res)
	})
	t.Run("exists", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		bs := make([]blocks.Block, 2)
		for i := range bs {
			bs[i] = testutil.NewRandBlock(10)
		}
		fx.index.EXPECT().FilterExistingOnly(ctx, testutil.BlocksToKeys(bs)).Return(testutil.BlocksToKeys(bs)[1:], nil)
		var resCh = make(chan blocks.Block)
		fx.store.EXPECT().GetMany(ctx, testutil.BlocksToKeys(bs[1:])).Return(resCh)
		res := fx.GetMany(ctx, testutil.BlocksToKeys(bs))
		assert.EqualValues(t, resCh, res)
	})
}

func TestServerStore_Delete(t *testing.T) {
	t.Run("unbind", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		b := testutil.NewRandBlock(10)
		var spaceId = "space1"
		sCtx := fileblockstore.CtxWithSpaceId(ctx, spaceId)
		fx.index.EXPECT().UnBind(sCtx, spaceId, []cid.Cid{b.Cid()}).Return(nil, nil)
		require.NoError(t, fx.Delete(sCtx, b.Cid()))
	})
	t.Run("unbind and delete", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		b := testutil.NewRandBlock(10)
		var spaceId = "space1"
		sCtx := fileblockstore.CtxWithSpaceId(ctx, spaceId)
		fx.index.EXPECT().UnBind(sCtx, spaceId, []cid.Cid{b.Cid()}).Return([]cid.Cid{b.Cid()}, nil)
		fx.store.EXPECT().DeleteMany(sCtx, []cid.Cid{b.Cid()})
		require.NoError(t, fx.Delete(sCtx, b.Cid()))
	})
}

func newFixture(t *testing.T) *fixture {
	ctrl := gomock.NewController(t)
	fx := &fixture{
		Service: New(),
		index:   mock_index.NewMockIndex(ctrl),
		store:   mock_serverstore.NewMockServerStore(ctrl),
		ctrl:    ctrl,
		a:       new(app.App),
	}

	fx.index.EXPECT().Name().Return(redisindex.CName).AnyTimes()
	fx.index.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.store.EXPECT().Name().Return(fileblockstore.CName).AnyTimes()
	fx.store.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.a.Register(fx.index).Register(fx.store).Register(fx.Service)
	require.NoError(t, fx.a.Start(ctx))
	return fx
}

type fixture struct {
	Service
	index *mock_index.MockIndex
	store *mock_serverstore.MockServerStore
	ctrl  *gomock.Controller
	a     *app.App
}

func (fx *fixture) Finish(t *testing.T) {
	fx.ctrl.Finish()
	require.NoError(t, fx.a.Close(ctx))
}
