package redisprovider

import (
	"context"
	"github.com/anytypeio/any-sync/app"
	"github.com/redis/go-redis/v9"
)

const CName = "filenode.redisprovider"

func New() RedisProvider {
	return new(redisProvider)
}

type RedisProvider interface {
	Redis() redis.UniversalClient
	app.ComponentRunnable
}

type redisProvider struct {
	redis redis.UniversalClient
}

func (r *redisProvider) Init(a *app.App) (err error) {
	conf := a.MustComponent("config").(configSource).GetRedis()
	if conf.IsCluster {
		opts, e := redis.ParseClusterURL(conf.Url)
		if e != nil {
			return e
		}
		r.redis = redis.NewClusterClient(opts)
	} else {
		opts, e := redis.ParseURL(conf.Url)
		if e != nil {
			return e
		}
		r.redis = redis.NewClient(opts)
	}
	return
}

func (r *redisProvider) Name() (name string) {
	return CName
}

func (r *redisProvider) Run(ctx context.Context) (err error) {
	return r.redis.Ping(ctx).Err()
}

func (r *redisProvider) Redis() redis.UniversalClient {
	return r.redis
}

func (r *redisProvider) Close(ctx context.Context) (err error) {
	if r.redis != nil {
		return r.redis.Close()
	}
	return
}
