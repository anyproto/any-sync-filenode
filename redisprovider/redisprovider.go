package redisprovider

import (
	"context"

	"github.com/anyproto/any-sync/app"
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
	if conf.Url == "" {
		conf.IsCluster = false
		conf.Url = "redis://127.0.0.1:6379/?dial_timeout=3&db=1&read_timeout=6s&max_retries=2"
	}
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
	if err = r.redis.Ping(ctx).Err(); err != nil {
		return
	}
	if err = r.redis.BFAdd(ctx, "_test_bf", 1).Err(); err != nil {
		return
	}
	return r.redis.Del(ctx, "_test_bf").Err()
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
