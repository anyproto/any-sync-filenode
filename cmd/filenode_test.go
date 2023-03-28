package main

import (
	"context"
	"github.com/anytypeio/any-sync-filenode/config"
	"github.com/anytypeio/any-sync-filenode/store/s3store"
	commonaccount "github.com/anytypeio/any-sync/accountservice"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/metric"
	"github.com/anytypeio/any-sync/net"
	"github.com/stretchr/testify/require"
	"testing"
)

var ctx = context.Background()

func TestBootstrap(t *testing.T) {
	c := &config.Config{
		Account: commonaccount.Config{
			PeerId:     "12D3KooWQxiZ5a7vcy4DTJa8Gy1eVUmwb5ojN4SrJC9Rjxzigw6C",
			PeerKey:    "X7YT92hRIQf42tNiIlOS01p1uJiPSqcdkE6LViS8PxnhAv/28YzTctqPjm11AC63Dq7ybmwPQiBu1pwa3AxHzQ==",
			SigningKey: "X7YT92hRIQf42tNiIlOS01p1uJiPSqcdkE6LViS8PxnhAv/28YzTctqPjm11AC63Dq7ybmwPQiBu1pwa3AxHzQ==",
		},
		GrpcServer: net.Config{},
		Metric:     metric.Config{},
		S3Store: s3store.Config{
			Bucket: "test",
		},
		FileDevStore: config.FileDevStore{},
	}
	a := new(app.App)
	a.Register(c)
	Bootstrap(a)
	require.NoError(t, a.Start(ctx))
	require.NoError(t, a.Close(ctx))
}
