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
	a := new(app.App)
	store := New()
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
