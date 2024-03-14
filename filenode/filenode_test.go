package filenode

import (
	"context"
	"testing"

	"github.com/anyproto/any-sync/acl"
	"github.com/anyproto/any-sync/acl/mock_acl"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/commonfile/fileblockstore"
	"github.com/anyproto/any-sync/commonfile/fileproto"
	"github.com/anyproto/any-sync/commonfile/fileproto/fileprotoerr"
	"github.com/anyproto/any-sync/metric"
	"github.com/anyproto/any-sync/net/peer"
	"github.com/anyproto/any-sync/net/rpc/server"
	"github.com/anyproto/any-sync/nodeconf"
	"github.com/anyproto/any-sync/nodeconf/mock_nodeconf"
	"github.com/anyproto/any-sync/util/crypto"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-sync-filenode/config"
	"github.com/anyproto/any-sync-filenode/index"
	"github.com/anyproto/any-sync-filenode/index/mock_index"
	"github.com/anyproto/any-sync-filenode/store/mock_store"
	"github.com/anyproto/any-sync-filenode/testutil"
)

func TestFileNode_Add(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		var (
			ctx, storeKey = newRandKey()
			fileId        = testutil.NewRandCid().String()
			b             = testutil.NewRandBlock(1024)
		)

		fx.aclService.EXPECT().OwnerPubKey(ctx, storeKey.SpaceId).Return(mustPubKey(ctx), nil)
		fx.index.EXPECT().CheckLimits(ctx, storeKey)
		fx.index.EXPECT().Migrate(ctx, storeKey)
		fx.index.EXPECT().BlocksLock(ctx, []blocks.Block{b}).Return(func() {}, nil)
		fx.index.EXPECT().BlocksGetNonExistent(ctx, []blocks.Block{b}).Return([]blocks.Block{b}, nil)
		fx.store.EXPECT().Add(ctx, []blocks.Block{b})
		fx.index.EXPECT().BlocksAdd(ctx, []blocks.Block{b})
		fx.index.EXPECT().CidEntriesByBlocks(ctx, []blocks.Block{b}).Return(&index.CidEntries{}, nil)
		fx.index.EXPECT().FileBind(ctx, storeKey, fileId, gomock.Any())
		fx.index.EXPECT().OnBlockUploaded(ctx, []blocks.Block{b})

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
			ctx, storeKey = newRandKey()
			fileId        = testutil.NewRandCid().String()
			b             = testutil.NewRandBlock(1024)
		)

		fx.aclService.EXPECT().OwnerPubKey(ctx, storeKey.SpaceId).Return(mustPubKey(ctx), nil)
		fx.index.EXPECT().Migrate(ctx, storeKey)
		fx.index.EXPECT().CheckLimits(ctx, storeKey).Return(index.ErrLimitExceed)

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
			ctx, storeKey = newRandKey()
			fileId        = testutil.NewRandCid().String()
			b             = testutil.NewRandBlock(1024)
			b2            = testutil.NewRandBlock(10)
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
			ctx, key = newRandKey()
			spaceId  = key.SpaceId
			fileId   = testutil.NewRandCid().String()
			b        = testutil.NewRandBlock(cidSizeLimit + 1)
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
		ctx, key := newRandKey()
		spaceId := key.SpaceId
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
		ctx, key := newRandKey()
		spaceId := key.SpaceId
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
	var ctx, storeKey = newRandKey()
	var bs = testutil.NewRandBlocks(3)
	cids := make([][]byte, len(bs))
	for _, b := range bs {
		cids = append(cids, b.Cid().Bytes())
	}
	fx.aclService.EXPECT().OwnerPubKey(ctx, storeKey.SpaceId).Return(mustPubKey(ctx), nil)
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
		ctx, storeKey = newRandKey()
		fileId        = testutil.NewRandCid().String()
		bs            = testutil.NewRandBlocks(3)
		cidsB         = make([][]byte, len(bs))
		cids          = make([]cid.Cid, len(bs))
		cidEntries    = &index.CidEntries{}
	)
	for i, b := range bs {
		cids[i] = b.Cid()
		cidsB[i] = b.Cid().Bytes()
	}

	fx.aclService.EXPECT().OwnerPubKey(ctx, storeKey.SpaceId).Return(mustPubKey(ctx), nil)
	fx.index.EXPECT().CheckLimits(ctx, storeKey)
	fx.index.EXPECT().Migrate(ctx, storeKey)
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
		ctx, storeKey = newRandKey()
		fileId1       = testutil.NewRandCid().String()
		fileId2       = testutil.NewRandCid().String()
	)
	fx.aclService.EXPECT().OwnerPubKey(ctx, storeKey.SpaceId).Return(mustPubKey(ctx), nil)
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

func TestFileNode_AccountInfo(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	var (
		ctx, storeKey = newRandKey()
		secondSpaceId = testutil.NewRandSpaceId()
	)
	fx.index.EXPECT().GroupInfo(ctx, storeKey.GroupId).Return(index.GroupInfo{
		BytesUsage:   100,
		CidsCount:    10,
		SpaceIds:     []string{storeKey.SpaceId, secondSpaceId},
		Limit:        90000,
		AccountLimit: 100000,
	}, nil)
	fx.index.EXPECT().SpaceInfo(ctx, storeKey).Return(index.SpaceInfo{
		BytesUsage: 90,
		CidsCount:  9,
		FileCount:  1,
	}, nil)
	fx.index.EXPECT().SpaceInfo(ctx, index.Key{GroupId: storeKey.GroupId, SpaceId: secondSpaceId}).Return(index.SpaceInfo{
		BytesUsage: 80,
		CidsCount:  8,
		FileCount:  2,
		Limit:      100,
	}, nil)

	resp, err := fx.handler.AccountInfo(ctx, &fileproto.AccountInfoRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Spaces, 2)
	assert.Equal(t, uint64(100), resp.TotalUsageBytes)
	assert.Equal(t, uint64(10), resp.TotalCidsCount)
	assert.Equal(t, uint64(90000), resp.LimitBytes)
	assert.Equal(t, uint64(100000), resp.AccountLimitBytes)

	assert.Equal(t, uint64(90), resp.Spaces[0].SpaceUsageBytes)
	assert.Equal(t, uint64(9), resp.Spaces[0].CidsCount)
	assert.Equal(t, uint64(1), resp.Spaces[0].FilesCount)
	assert.Equal(t, uint64(90000), resp.Spaces[0].LimitBytes)
	assert.Equal(t, uint64(100), resp.Spaces[0].TotalUsageBytes)

	assert.Equal(t, uint64(80), resp.Spaces[1].SpaceUsageBytes)
	assert.Equal(t, uint64(8), resp.Spaces[1].CidsCount)
	assert.Equal(t, uint64(2), resp.Spaces[1].FilesCount)
	assert.Equal(t, uint64(100), resp.Spaces[1].LimitBytes)
	assert.Equal(t, uint64(80), resp.Spaces[1].TotalUsageBytes)
}

func TestFileNode_SpaceInfo(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	var (
		ctx, storeKey = newRandKey()
	)
	fx.aclService.EXPECT().OwnerPubKey(ctx, storeKey.SpaceId).Return(mustPubKey(ctx), nil)
	fx.index.EXPECT().Migrate(ctx, storeKey)

	fx.index.EXPECT().GroupInfo(ctx, storeKey.GroupId).Return(index.GroupInfo{
		BytesUsage:   100,
		CidsCount:    10,
		SpaceIds:     []string{storeKey.SpaceId},
		Limit:        90000,
		AccountLimit: 100000,
	}, nil)
	fx.index.EXPECT().SpaceInfo(ctx, storeKey).Return(index.SpaceInfo{
		BytesUsage: 90,
		CidsCount:  9,
		FileCount:  1,
	}, nil)

	info, err := fx.SpaceInfo(ctx, storeKey.SpaceId)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, uint64(90), info.SpaceUsageBytes)
	assert.Equal(t, uint64(9), info.CidsCount)
	assert.Equal(t, uint64(1), info.FilesCount)
	assert.Equal(t, uint64(90000), info.LimitBytes)
	assert.Equal(t, uint64(100), info.TotalUsageBytes)
}

func TestFileNode_AccountLimitSet(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	var peerId = "peerId"

	fx.nodeConf.EXPECT().NodeTypes(peerId).Return([]nodeconf.NodeType{nodeconf.NodeTypeCoordinator})

	ctx := peer.CtxWithPeerId(context.Background(), peerId)

	_, storeKey := newRandKey()
	identity := storeKey.GroupId

	fx.index.EXPECT().SetGroupLimit(ctx, identity, uint64(12345))

	require.NoError(t, fx.AccountLimitSet(ctx, identity, 12345))
}

func TestFileNode_SpaceLimitSet(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	ctx, storeKey := newRandKey()

	fx.aclService.EXPECT().OwnerPubKey(ctx, storeKey.SpaceId).Return(mustPubKey(ctx), nil)
	fx.index.EXPECT().Migrate(ctx, storeKey)
	fx.index.EXPECT().SetSpaceLimit(ctx, storeKey, uint64(12345))
	require.NoError(t, fx.SpaceLimitSet(ctx, storeKey.SpaceId, 12345))
}

func newFixture(t *testing.T) *fixture {
	ctrl := gomock.NewController(t)
	fx := &fixture{
		fileNode:   New().(*fileNode),
		index:      mock_index.NewMockIndex(ctrl),
		store:      mock_store.NewMockStore(ctrl),
		aclService: mock_acl.NewMockAclService(ctrl),
		nodeConf:   mock_nodeconf.NewMockService(ctrl),
		serv:       server.New(),
		ctrl:       ctrl,
		a:          new(app.App),
	}

	fx.index.EXPECT().Name().Return(index.CName).AnyTimes()
	fx.index.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.index.EXPECT().Run(gomock.Any()).AnyTimes()
	fx.index.EXPECT().Close(gomock.Any()).AnyTimes()

	fx.store.EXPECT().Name().Return(fileblockstore.CName).AnyTimes()
	fx.store.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.aclService.EXPECT().Name().Return(acl.CName).AnyTimes()
	fx.aclService.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.aclService.EXPECT().Run(gomock.Any()).AnyTimes()
	fx.aclService.EXPECT().Close(gomock.Any()).AnyTimes()

	fx.nodeConf.EXPECT().Name().Return(nodeconf.CName).AnyTimes()
	fx.nodeConf.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.nodeConf.EXPECT().Run(gomock.Any()).AnyTimes()
	fx.nodeConf.EXPECT().Close(gomock.Any()).AnyTimes()

	fx.a.Register(metric.New()).
		Register(fx.serv).
		Register(fx.index).
		Register(fx.store).
		Register(fx.aclService).
		Register(fx.fileNode).
		Register(fx.nodeConf).
		Register(&config.Config{})
	require.NoError(t, fx.a.Start(context.Background()))
	return fx
}

type fixture struct {
	*fileNode
	index      *mock_index.MockIndex
	store      *mock_store.MockStore
	ctrl       *gomock.Controller
	a          *app.App
	aclService *mock_acl.MockAclService
	serv       server.DRPCServer
	nodeConf   *mock_nodeconf.MockService
}

func (fx *fixture) Finish(t *testing.T) {
	fx.ctrl.Finish()
	require.NoError(t, fx.a.Close(context.Background()))
}

func newRandKey() (context.Context, index.Key) {
	_, pubKey, _ := crypto.GenerateRandomEd25519KeyPair()
	pubKeyRaw, _ := pubKey.Marshall()
	return peer.CtxWithIdentity(context.Background(), pubKeyRaw), index.Key{
		SpaceId: testutil.NewRandSpaceId(),
		GroupId: pubKey.Account(),
	}
}

func mustPubKey(ctx context.Context) crypto.PubKey {
	pubKey, err := peer.CtxPubKey(ctx)
	if err != nil {
		panic(err)
	}
	return pubKey
}
