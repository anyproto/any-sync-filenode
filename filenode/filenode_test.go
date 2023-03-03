package filenode

import (
	"context"
	"github.com/anytypeio/any-sync-filenode/index"
	"github.com/anytypeio/any-sync-filenode/index/mock_index"
	"github.com/anytypeio/any-sync-filenode/index/redisindex"
	"github.com/anytypeio/any-sync-filenode/limit"
	"github.com/anytypeio/any-sync-filenode/limit/mock_limit"
	"github.com/anytypeio/any-sync-filenode/store/mock_store"
	"github.com/anytypeio/any-sync-filenode/testutil"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/commonfile/fileblockstore"
	"github.com/anytypeio/any-sync/commonfile/fileproto"
	"github.com/anytypeio/any-sync/net/rpc/rpctest"
	"github.com/golang/mock/gomock"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-libipfs/blocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var ctx = context.Background()

func TestFileNode_Add(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		var (
			spaceId = testutil.NewRandSpaceId()
			fileId  = testutil.NewRandCid().String()
			b       = testutil.NewRandBlock(1024)
		)

		fx.limit.EXPECT().Check(ctx, spaceId).Return(uint64(123), nil)
		fx.index.EXPECT().SpaceSize(ctx, spaceId).Return(uint64(120), nil)
		fx.index.EXPECT().Lock(ctx, []cid.Cid{b.Cid()}).Return(func() {}, nil)
		fx.index.EXPECT().GetNonExistentBlocks(ctx, []blocks.Block{b}).Return([]blocks.Block{b}, nil)
		fx.store.EXPECT().Add(ctx, []blocks.Block{b})
		fx.index.EXPECT().Bind(ctx, spaceId, fileId, []blocks.Block{b})

		resp, err := fx.handler.BlockPush(ctx, &fileproto.BlockPushRequest{
			SpaceId: spaceId,
			FileId:  fileId,
			Cid:     b.Cid().Bytes(),
			Data:    b.RawData(),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
	t.Run("limit exceed", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		var (
			spaceId = testutil.NewRandSpaceId()
			fileId  = testutil.NewRandCid().String()
			b       = testutil.NewRandBlock(1024)
		)

		fx.limit.EXPECT().Check(ctx, spaceId).Return(uint64(123), nil)
		fx.index.EXPECT().SpaceSize(ctx, spaceId).Return(uint64(124), nil)

		resp, err := fx.handler.BlockPush(ctx, &fileproto.BlockPushRequest{
			SpaceId: spaceId,
			FileId:  fileId,
			Cid:     b.Cid().Bytes(),
			Data:    b.RawData(),
		})
		require.EqualError(t, err, ErrLimitExceed.Error())
		require.Nil(t, resp)
	})
	t.Run("invalid cid", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		var (
			spaceId = testutil.NewRandSpaceId()
			fileId  = testutil.NewRandCid().String()
			b       = testutil.NewRandBlock(1024)
			b2      = testutil.NewRandBlock(10)
		)

		resp, err := fx.handler.BlockPush(ctx, &fileproto.BlockPushRequest{
			SpaceId: spaceId,
			FileId:  fileId,
			Cid:     b2.Cid().Bytes(),
			Data:    b.RawData(),
		})
		require.EqualError(t, err, ErrWrongHash.Error())
		assert.Nil(t, resp)
	})
	t.Run("err data too bif", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		var (
			spaceId = testutil.NewRandSpaceId()
			fileId  = testutil.NewRandCid().String()
			b       = testutil.NewRandBlock(cidSizeLimit + 1)
		)

		resp, err := fx.handler.BlockPush(ctx, &fileproto.BlockPushRequest{
			SpaceId: spaceId,
			FileId:  fileId,
			Cid:     b.Cid().Bytes(),
			Data:    b.RawData(),
		})
		require.EqualError(t, err, ErrCidDataTooBig.Error())
		assert.Nil(t, resp)
	})
}

func TestFileNode_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		spaceId := testutil.NewRandSpaceId()
		b := testutil.NewRandBlock(10)
		fx.index.EXPECT().Exists(gomock.Any(), b.Cid()).Return(true, nil)
		fx.store.EXPECT().Get(ctx, b.Cid()).Return(b, nil)
		resp, err := fx.handler.BlockGet(ctx, &fileproto.BlockGetRequest{
			SpaceId: spaceId,
			Cid:     b.Cid().Bytes(),
		})
		require.NoError(t, err)
		assert.Equal(t, b.RawData(), resp.Data)
	})
	t.Run("not found", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		spaceId := testutil.NewRandSpaceId()
		b := testutil.NewRandBlock(10)
		fx.index.EXPECT().Exists(gomock.Any(), b.Cid()).Return(false, nil)
		resp, err := fx.handler.BlockGet(ctx, &fileproto.BlockGetRequest{
			SpaceId: spaceId,
			Cid:     b.Cid().Bytes(),
		})
		require.EqualError(t, err, fileblockstore.ErrCIDNotFound.Error())
		assert.Nil(t, resp)
	})
}

func TestFileNode_Check(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	var spaceId = testutil.NewRandSpaceId()
	var bs = testutil.NewRandBlocks(3)
	cids := make([][]byte, len(bs))
	for _, b := range bs {
		cids = append(cids, b.Cid().Bytes())
	}
	fx.limit.EXPECT().Check(ctx, spaceId)
	fx.index.EXPECT().ExistsInSpace(ctx, spaceId, testutil.BlocksToKeys(bs)).Return(testutil.BlocksToKeys(bs[:1]), nil)
	fx.index.EXPECT().Exists(ctx, bs[1].Cid()).Return(true, nil)
	fx.index.EXPECT().Exists(ctx, bs[2].Cid()).Return(false, nil)
	resp, err := fx.handler.BlocksCheck(ctx, &fileproto.BlocksCheckRequest{
		SpaceId: spaceId,
		Cids:    cids,
	})
	require.NoError(t, err)
	require.Len(t, resp.BlocksAvailability, len(bs))
	assert.Equal(t, fileproto.AvailabilityStatus_ExistsInSpace, resp.BlocksAvailability[0].Status)
	assert.Equal(t, fileproto.AvailabilityStatus_Exists, resp.BlocksAvailability[1].Status)
	assert.Equal(t, fileproto.AvailabilityStatus_NotExists, resp.BlocksAvailability[2].Status)
}

func TestFileNode_BlocksBind(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	var (
		spaceId = testutil.NewRandSpaceId()
		fileId  = testutil.NewRandCid().String()
		bs      = testutil.NewRandBlocks(3)
		cidsB   = make([][]byte, len(bs))
		cids    = make([]cid.Cid, len(bs))
	)
	for i, b := range bs {
		cids[i] = b.Cid()
		cidsB[i] = b.Cid().Bytes()
	}

	fx.limit.EXPECT().Check(ctx, spaceId).Return(uint64(123), nil)
	fx.index.EXPECT().SpaceSize(ctx, spaceId).Return(uint64(12), nil)
	fx.index.EXPECT().Lock(ctx, cids).Return(func() {}, nil)
	fx.index.EXPECT().BindCids(ctx, spaceId, fileId, cids)

	resp, err := fx.handler.BlocksBind(ctx, &fileproto.BlocksBindRequest{
		SpaceId: spaceId,
		FileId:  fileId,
		Cids:    cidsB,
	})
	require.NotNil(t, resp)
	require.NoError(t, err)
}

func TestFileNode_FileInfo(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	var (
		spaceId = testutil.NewRandSpaceId()
		fileId1 = testutil.NewRandCid().String()
		fileId2 = testutil.NewRandCid().String()
	)
	fx.limit.EXPECT().Check(ctx, spaceId).AnyTimes()
	fx.index.EXPECT().FileInfo(ctx, spaceId, fileId1).Return(index.FileInfo{1, 1}, nil)
	fx.index.EXPECT().FileInfo(ctx, spaceId, fileId2).Return(index.FileInfo{2, 2}, nil)

	resp, err := fx.handler.FilesInfo(ctx, &fileproto.FilesInfoRequest{
		SpaceId: spaceId,
		FileIds: []string{fileId1, fileId2},
	})
	require.NoError(t, err)
	require.Len(t, resp.FilesInfo, 2)
	assert.Equal(t, uint64(1), resp.FilesInfo[0].UsageBytes)
	assert.Equal(t, uint64(2), resp.FilesInfo[1].UsageBytes)
}

func newFixture(t *testing.T) *fixture {
	ctrl := gomock.NewController(t)
	fx := &fixture{
		fileNode: New().(*fileNode),
		index:    mock_index.NewMockIndex(ctrl),
		store:    mock_store.NewMockStore(ctrl),
		limit:    mock_limit.NewMockLimit(ctrl),
		serv:     rpctest.NewTestServer(),
		ctrl:     ctrl,
		a:        new(app.App),
	}

	fx.index.EXPECT().Name().Return(redisindex.CName).AnyTimes()
	fx.index.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.store.EXPECT().Name().Return(fileblockstore.CName).AnyTimes()
	fx.store.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.limit.EXPECT().Name().Return(limit.CName).AnyTimes()
	fx.limit.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.limit.EXPECT().Run(gomock.Any()).AnyTimes()
	fx.limit.EXPECT().Close(gomock.Any()).AnyTimes()

	fx.a.Register(fx.serv).Register(fx.index).Register(fx.store).Register(fx.limit).Register(fx.fileNode)
	require.NoError(t, fx.a.Start(ctx))
	return fx
}

type fixture struct {
	*fileNode
	index *mock_index.MockIndex
	store *mock_store.MockStore
	ctrl  *gomock.Controller
	a     *app.App
	limit *mock_limit.MockLimit
	serv  *rpctest.TesServer
}

func (fx *fixture) Finish(t *testing.T) {
	fx.ctrl.Finish()
	require.NoError(t, fx.a.Close(ctx))
}
