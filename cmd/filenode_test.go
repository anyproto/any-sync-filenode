package main

import (
	"context"
	"os"
	"testing"

	commonaccount "github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/metric"
	"github.com/anyproto/any-sync/net/rpc"
	"github.com/stretchr/testify/require"

	"github.com/anyproto/any-sync-filenode/config"
	"github.com/anyproto/any-sync-filenode/store/s3store"
)

var ctx = context.Background()

func TestBootstrap(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	c := &config.Config{
		Account: commonaccount.Config{
			PeerId:     "12D3KooWQxiZ5a7vcy4DTJa8Gy1eVUmwb5ojN4SrJC9Rjxzigw6C",
			PeerKey:    "X7YT92hRIQf42tNiIlOS01p1uJiPSqcdkE6LViS8PxnhAv/28YzTctqPjm11AC63Dq7ybmwPQiBu1pwa3AxHzQ==",
			SigningKey: "X7YT92hRIQf42tNiIlOS01p1uJiPSqcdkE6LViS8PxnhAv/28YzTctqPjm11AC63Dq7ybmwPQiBu1pwa3AxHzQ==",
		},
		Drpc:   rpc.Config{},
		Metric: metric.Config{},
		S3Store: s3store.Config{
			Bucket:      "test",
			IndexBucket: "testIndex",
		},
		FileDevStore:     config.FileDevStore{},
		NetworkStorePath: tmpDir,
	}
	a := new(app.App)
	a.Register(c)
	Bootstrap(a)
	require.NoError(t, a.Start(ctx))
	require.NoError(t, a.Close(ctx))
	_ = os.RemoveAll(tmpDir)
}
