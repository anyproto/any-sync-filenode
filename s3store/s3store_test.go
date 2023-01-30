package s3store

import (
	"context"
	"fmt"
	"github.com/anytypeio/any-sync/app"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var ctx = context.Background()

func TestS3store_GetMany(t *testing.T) {
	// skip the test because it needs amazon credentials
	t.Skip()

	a := new(app.App)
	store := New()
	a.Register(&config{})
	a.Register(store)
	require.NoError(t, a.Start(ctx))
	defer a.Close(ctx)

	var bs []blocks.Block
	var ks []cid.Cid
	for i := 0; i < 10; i++ {
		b := blocks.NewBlock([]byte(fmt.Sprint(i)))
		bs = append(bs, b)
		ks = append(ks, b.Cid())
	}
	require.NoError(t, store.Add(ctx, bs))

	ctxGetMany, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	res := store.GetMany(ctxGetMany, ks)
	var result []blocks.Block
	for b := range res {
		result = append(result, b)
	}
	require.NoError(t, ctx.Err())
	require.Len(t, result, 10)

	for _, k := range ks {
		assert.NoError(t, store.Delete(ctx, k))
	}
}

func TestS3store_Check(t *testing.T) {
	// skip the test because it needs amazon credentials
	t.Skip()

	a := new(app.App)
	store := New()
	a.Register(&config{})
	a.Register(store)
	require.NoError(t, a.Start(ctx))
	defer a.Close(ctx)

	var bs []blocks.Block
	var ks []cid.Cid
	for i := 0; i < 5; i++ {
		b := blocks.NewBlock([]byte(fmt.Sprint("checkTest", i)))
		if i < 3 {
			bs = append(bs, b)
		}
		ks = append(ks, b.Cid())
	}
	require.NoError(t, store.Add(ctx, bs))

	result, err := store.Check(ctx, "", ks...)
	require.NoError(t, err)
	for _, r := range result {
		t.Log(r.Cid, r.Status)
	}

	for _, b := range bs {
		assert.NoError(t, store.Delete(ctx, b.Cid()))
	}
}

type config struct {
}

func (c config) Init(a *app.App) error { return nil }
func (c config) Name() string          { return "config" }

func (c config) GetS3Store() Config {
	return Config{
		Region:     "eu-central-1",
		Bucket:     "anytype-test",
		MaxThreads: 4,
	}
}
