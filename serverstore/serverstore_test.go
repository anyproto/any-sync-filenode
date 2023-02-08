package serverstore

import (
	"context"
	"github.com/anytypeio/any-sync-filenode/index/mock_index"
	"github.com/anytypeio/any-sync-filenode/index/redisindex"
	"github.com/anytypeio/any-sync-filenode/serverstore/mock_serverstore"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/commonfile/fileblockstore"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
)

var ctx = context.Background()

func TestServerStore_Add(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
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
