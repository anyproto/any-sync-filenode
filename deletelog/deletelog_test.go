package deletelog

import (
	"context"
	"testing"
	"time"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/coordinator/coordinatorclient"
	"github.com/anyproto/any-sync/coordinator/coordinatorclient/mock_coordinatorclient"
	"github.com/anyproto/any-sync/coordinator/coordinatorproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-sync-filenode/index"
	"github.com/anyproto/any-sync-filenode/index/mock_index"
	"github.com/anyproto/any-sync-filenode/redisprovider/testredisprovider"
)

var ctx = context.Background()

func TestDeleteLog_checkLog(t *testing.T) {
	t.Run("no records", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		fx.coord.EXPECT().DeletionLog(ctx, "", recordsLimit).Return(nil, nil)
		require.NoError(t, fx.checkLog(ctx))
	})
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		now := time.Now().Unix()
		fx.coord.EXPECT().DeletionLog(ctx, "", recordsLimit).Return([]*coordinatorproto.DeletionLogRecord{
			{
				Id:        "1",
				SpaceId:   "s1",
				Status:    coordinatorproto.DeletionLogRecordStatus_Ok,
				Timestamp: now,
				FileGroup: "f1",
			},
			{
				Id:        "2",
				SpaceId:   "s2",
				Status:    coordinatorproto.DeletionLogRecordStatus_Remove,
				Timestamp: now,
				FileGroup: "f2",
			},
		}, nil)
		fx.index.EXPECT().SpaceDelete(ctx, index.Key{
			GroupId: "f2",
			SpaceId: "s2",
		})
		require.NoError(t, fx.checkLog(ctx))
		lastId, err := fx.redis.Get(ctx, lastKey).Result()
		require.NoError(t, err)
		assert.Equal(t, "2", lastId)
	})

}

func newFixture(t *testing.T) *fixture {
	ctrl := gomock.NewController(t)
	fx := &fixture{
		ctrl:      ctrl,
		a:         new(app.App),
		coord:     mock_coordinatorclient.NewMockCoordinatorClient(ctrl),
		index:     mock_index.NewMockIndex(ctrl),
		deleteLog: New().(*deleteLog),
	}
	fx.disableTicker = true
	fx.coord.EXPECT().Name().Return(coordinatorclient.CName).AnyTimes()
	fx.coord.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.index.EXPECT().Name().Return(index.CName).AnyTimes()
	fx.index.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.index.EXPECT().Run(gomock.Any()).AnyTimes()
	fx.index.EXPECT().Close(gomock.Any()).AnyTimes()

	fx.a.Register(testredisprovider.NewTestRedisProviderNum(7)).Register(fx.coord).Register(fx.index).Register(fx.deleteLog)
	require.NoError(t, fx.a.Start(ctx))

	return fx
}

type fixture struct {
	ctrl  *gomock.Controller
	a     *app.App
	coord *mock_coordinatorclient.MockCoordinatorClient
	index *mock_index.MockIndex
	*deleteLog
}

func (fx *fixture) finish(t *testing.T) {
	require.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}
