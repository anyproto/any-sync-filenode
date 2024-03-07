package index

import (
	"context"
	"testing"

	"github.com/anyproto/any-sync/app"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-sync-filenode/config"
	"github.com/anyproto/any-sync-filenode/redisprovider/testredisprovider"
	"github.com/anyproto/any-sync-filenode/store/mock_store"
	"github.com/anyproto/any-sync-filenode/store/s3store"
	"github.com/anyproto/any-sync-filenode/testutil"
)

var ctx = context.Background()

func newRandKey() Key {
	return Key{
		SpaceId: testutil.NewRandSpaceId(),
		GroupId: "A" + testutil.NewRandCid().String(),
	}
}

func newFixture(t *testing.T) (fx *fixture) {
	ctrl := gomock.NewController(t)
	fx = &fixture{
		redisIndex:   New().(*redisIndex),
		ctrl:         ctrl,
		persistStore: mock_store.NewMockStore(ctrl),
		a:            new(app.App),
	}
	fx.persistStore.EXPECT().Name().Return(s3store.CName).AnyTimes()
	fx.persistStore.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.a.Register(testredisprovider.NewTestRedisProvider()).Register(fx.redisIndex).Register(fx.persistStore).Register(&config.Config{DefaultLimit: 1024})
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
