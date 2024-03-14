package index

import (
	"context"
	"testing"

	"github.com/anyproto/any-sync/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-sync-filenode/config"
	"github.com/anyproto/any-sync-filenode/redisprovider/testredisprovider"
	"github.com/anyproto/any-sync-filenode/store/mock_store"
	"github.com/anyproto/any-sync-filenode/store/s3store"
	"github.com/anyproto/any-sync-filenode/testutil"
)

var ctx = context.Background()

func TestRedisIndex_FilesList(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	k := newRandKey()
	limit := uint64(10)
	require.NoError(t, fx.SetSpaceLimit(ctx, k, limit))
	// no error
	assert.NoError(t, fx.CheckLimits(ctx, k))

	bs := testutil.NewRandBlocks(3)
	require.NoError(t, fx.BlocksAdd(ctx, bs))
	fileId := testutil.NewRandCid().String()
	cids, err := fx.CidEntriesByBlocks(ctx, bs)
	require.NoError(t, err)
	require.NoError(t, fx.FileBind(ctx, k, fileId, cids))
	cids.Release()

	fileIds, err := fx.FilesList(ctx, k)
	require.NoError(t, err)
	assert.Equal(t, []string{fileId}, fileIds)
}

func newRandKey() Key {
	return Key{
		SpaceId: testutil.NewRandSpaceId(),
		GroupId: "A" + testutil.NewRandCid().String(),
	}
}

func newFixture(t *testing.T) (fx *fixture) {
	return newFixtureConfig(t, nil)
}

func newFixtureConfig(t *testing.T, conf *config.Config) (fx *fixture) {
	ctrl := gomock.NewController(t)
	fx = &fixture{
		redisIndex:   New().(*redisIndex),
		ctrl:         ctrl,
		persistStore: mock_store.NewMockStore(ctrl),
		a:            new(app.App),
	}
	fx.persistStore.EXPECT().Name().Return(s3store.CName).AnyTimes()
	fx.persistStore.EXPECT().Init(gomock.Any()).AnyTimes()
	if conf == nil {
		conf = &config.Config{DefaultLimit: 1024, PersistTtl: 3600}
	}
	fx.a.Register(testredisprovider.NewTestRedisProvider()).Register(fx.redisIndex).Register(fx.persistStore).Register(conf)
	require.NoError(t, fx.a.Start(ctx))
	return
}

type fixture struct {
	*redisIndex
	a            *app.App
	ctrl         *gomock.Controller
	persistStore *mock_store.MockStore
}

func (fx *fixture) Finish(t require.TestingT) {
	require.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}
