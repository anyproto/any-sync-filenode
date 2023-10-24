package filenode

import (
	"context"
	"testing"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/commonfile/fileblockstore"
	"github.com/anyproto/any-sync/commonfile/fileproto"
	"github.com/anyproto/any-sync/commonfile/fileproto/fileprotoerr"
	"github.com/anyproto/any-sync/metric"
	"github.com/anyproto/any-sync/net/rpc/server"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-sync-filenode/config"
	"github.com/anyproto/any-sync-filenode/index"
	"github.com/anyproto/any-sync-filenode/index/mock_index"
	"github.com/anyproto/any-sync-filenode/limit"
	"github.com/anyproto/any-sync-filenode/limit/mock_limit"
	"github.com/anyproto/any-sync-filenode/store/mock_store"
	"github.com/anyproto/any-sync-filenode/testutil"
)

var ctx = context.Background()

func TestFileNode_Add(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		var (
			storeKey = newRandKey()
			fileId   = testutil.NewRandCid().String()
			b        = testutil.NewRandBlock(1024)
		)

		fx.limit.EXPECT().Check(ctx, storeKey.SpaceId).Return(uint64(123), storeKey.GroupId, nil)
		fx.index.EXPECT().Migrate(ctx, storeKey)
		fx.index.EXPECT().GroupInfo(ctx, storeKey.GroupId).Return(index.GroupInfo{BytesUsage: uint64(120)}, nil)
		fx.index.EXPECT().BlocksLock(ctx, []blocks.Block{b}).Return(func() {}, nil)
		fx.index.EXPECT().BlocksGetNonExistent(ctx, []blocks.Block{b}).Return([]blocks.Block{b}, nil)
		fx.store.EXPECT().Add(ctx, []blocks.Block{b})
		fx.index.EXPECT().BlocksAdd(ctx, []blocks.Block{b})
		fx.index.EXPECT().CidEntriesByBlocks(ctx, []blocks.Block{b}).Return(&index.CidEntries{}, nil)
		fx.index.EXPECT().FileBind(ctx, storeKey, fileId, gomock.Any())

		resp, err := fx.handler.BlockPush(ctx, &fileproto.BlockPushRequest{
			SpaceId: storeKey.SpaceId,
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
			storeKey = newRandKey()
			fileId   = testutil.NewRandCid().String()
			b        = testutil.NewRandBlock(1024)
		)

		fx.limit.EXPECT().Check(ctx, storeKey.SpaceId).Return(uint64(123), storeKey.GroupId, nil)
		fx.index.EXPECT().Migrate(ctx, storeKey)
		fx.index.EXPECT().GroupInfo(ctx, storeKey.GroupId).Return(index.GroupInfo{BytesUsage: uint64(124)}, nil)

		resp, err := fx.handler.BlockPush(ctx, &fileproto.BlockPushRequest{
			SpaceId: storeKey.SpaceId,
			FileId:  fileId,
			Cid:     b.Cid().Bytes(),
			Data:    b.RawData(),
		})
		require.EqualError(t, err, fileprotoerr.ErrSpaceLimitExceeded.Error())
		require.Nil(t, resp)
	})
	t.Run("invalid cid", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		var (
			storeKey = newRandKey()
			fileId   = testutil.NewRandCid().String()
			b        = testutil.NewRandBlock(1024)
			b2       = testutil.NewRandBlock(10)
		)

		resp, err := fx.handler.BlockPush(ctx, &fileproto.BlockPushRequest{
			SpaceId: storeKey.SpaceId,
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
		require.EqualError(t, err, fileprotoerr.ErrQuerySizeExceeded.Error())
		assert.Nil(t, resp)
	})

}

func TestFileNode_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		spaceId := testutil.NewRandSpaceId()
		b := testutil.NewRandBlock(10)
		fx.index.EXPECT().CidExists(gomock.Any(), b.Cid()).Return(true, nil)
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
		fx.index.EXPECT().CidExists(gomock.Any(), b.Cid()).Return(false, nil)
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
	var storeKey = newRandKey()
	var bs = testutil.NewRandBlocks(3)
	cids := make([][]byte, len(bs))
	for _, b := range bs {
		cids = append(cids, b.Cid().Bytes())
	}
	fx.limit.EXPECT().Check(ctx, storeKey.SpaceId).Return(uint64(100000), storeKey.GroupId, nil)
	fx.index.EXPECT().Migrate(ctx, storeKey)
	fx.index.EXPECT().CidExistsInSpace(ctx, storeKey, testutil.BlocksToKeys(bs)).Return(testutil.BlocksToKeys(bs[:1]), nil)
	fx.index.EXPECT().CidExists(ctx, bs[1].Cid()).Return(true, nil)
	fx.index.EXPECT().CidExists(ctx, bs[2].Cid()).Return(false, nil)
	resp, err := fx.handler.BlocksCheck(ctx, &fileproto.BlocksCheckRequest{
		SpaceId: storeKey.SpaceId,
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
		storeKey   = newRandKey()
		fileId     = testutil.NewRandCid().String()
		bs         = testutil.NewRandBlocks(3)
		cidsB      = make([][]byte, len(bs))
		cids       = make([]cid.Cid, len(bs))
		cidEntries = &index.CidEntries{}
	)
	for i, b := range bs {
		cids[i] = b.Cid()
		cidsB[i] = b.Cid().Bytes()
	}

	fx.limit.EXPECT().Check(ctx, storeKey.SpaceId).Return(uint64(123), storeKey.GroupId, nil)
	fx.index.EXPECT().Migrate(ctx, storeKey)
	fx.index.EXPECT().GroupInfo(ctx, storeKey.GroupId).Return(index.GroupInfo{BytesUsage: 12}, nil)
	fx.index.EXPECT().CidEntries(ctx, cids).Return(cidEntries, nil)
	fx.index.EXPECT().FileBind(ctx, storeKey, fileId, cidEntries)

	resp, err := fx.handler.BlocksBind(ctx, &fileproto.BlocksBindRequest{
		SpaceId: storeKey.SpaceId,
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
		storeKey = newRandKey()
		fileId1  = testutil.NewRandCid().String()
		fileId2  = testutil.NewRandCid().String()
	)
	fx.limit.EXPECT().Check(ctx, storeKey.SpaceId).AnyTimes().Return(uint64(100000), storeKey.GroupId, nil)
	fx.index.EXPECT().Migrate(ctx, storeKey)
	fx.index.EXPECT().FileInfo(ctx, storeKey, fileId1, fileId2).Return([]index.FileInfo{{1, 1}, {2, 2}}, nil)

	resp, err := fx.handler.FilesInfo(ctx, &fileproto.FilesInfoRequest{
		SpaceId: storeKey.SpaceId,
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
		serv:     server.New(),
		ctrl:     ctrl,
		a:        new(app.App),
	}

	fx.index.EXPECT().Name().Return(index.CName).AnyTimes()
	fx.index.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.index.EXPECT().Run(gomock.Any()).AnyTimes()
	fx.index.EXPECT().Close(gomock.Any()).AnyTimes()

	fx.store.EXPECT().Name().Return(fileblockstore.CName).AnyTimes()
	fx.store.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.limit.EXPECT().Name().Return(limit.CName).AnyTimes()
	fx.limit.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.limit.EXPECT().Run(gomock.Any()).AnyTimes()
	fx.limit.EXPECT().Close(gomock.Any()).AnyTimes()

	fx.a.Register(metric.New()).Register(fx.serv).Register(fx.index).Register(fx.store).Register(fx.limit).Register(fx.fileNode).Register(&config.Config{})
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
	serv  server.DRPCServer
}

func (fx *fixture) Finish(t *testing.T) {
	fx.ctrl.Finish()
	require.NoError(t, fx.a.Close(ctx))
}

func newRandKey() index.Key {
	return index.Key{
		SpaceId: testutil.NewRandSpaceId(),
		GroupId: "A" + testutil.NewRandCid().String(),
	}
}
