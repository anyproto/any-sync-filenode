package testredisprovider

import (
	"context"
	"fmt"

	"github.com/anyproto/any-sync/app"
	"github.com/redis/go-redis/v9"

	"github.com/anyproto/any-sync-filenode/redisprovider"
)

func NewTestRedisProvider() *TestRedisProvider {
	return &TestRedisProvider{
		db:           6,
		flushOnStart: true,
	}
}

func NewTestRedisProviderNum(num int) *TestRedisProvider {
	return &TestRedisProvider{
		db:           num,
		flushOnStart: true,
	}
}

type TestRedisProvider struct {
	client       redis.UniversalClient
	flushOnStart bool
	db           int
}

func (t *TestRedisProvider) Init(a *app.App) (err error) {
	opts, err := redis.ParseURL(fmt.Sprintf("redis://127.0.0.1:6379/?dial_timeout=3&db=1&read_timeout=6s&max_retries=2&db=%d", t.db))
	if err != nil {
		return
	}
	t.client = redis.NewClient(opts)
	return nil
}

func (t *TestRedisProvider) Name() (name string) {
	return redisprovider.CName
}

func (t *TestRedisProvider) Redis() redis.UniversalClient {
	return t.client
}

func (t *TestRedisProvider) Run(ctx context.Context) (err error) {
	if t.flushOnStart {
		return t.client.FlushDB(ctx).Err()
	}
	return t.client.Ping(ctx).Err()
}

func (t *TestRedisProvider) Close(ctx context.Context) (err error) {
	return t.client.Close()
}

func (t *TestRedisProvider) WithDb(db int) *TestRedisProvider {
	t.db = db
	return t
}

func (t *TestRedisProvider) WithFLush(flushOnStart bool) *TestRedisProvider {
	t.flushOnStart = flushOnStart
	return t
}
