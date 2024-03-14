package index

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anyproto/any-sync-filenode/testutil"
)

func TestRedisIndex_WaitCidExists(t *testing.T) {
	t.Run("cid exists", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)

		k := newRandKey()
		bs := testutil.NewRandBlocks(1)
		require.NoError(t, fx.BlocksAdd(ctx, bs))
		fileId := testutil.NewRandCid().String()
		cids, err := fx.CidEntriesByBlocks(ctx, bs)
		require.NoError(t, err)
		require.NoError(t, fx.FileBind(ctx, k, fileId, cids))
		cids.Release()

		tCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		require.NoError(t, fx.WaitCidExists(tCtx, bs[0].Cid()))
	})

	t.Run("cid wait", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		bs := testutil.NewRandBlocks(2)

		wg := &sync.WaitGroup{}
		wg.Add(3)
		tCtx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		go func() {
			defer wg.Done()
			require.NoError(t, fx.WaitCidExists(tCtx, bs[0].Cid()))
		}()
		go func() {
			defer wg.Done()
			require.NoError(t, fx.WaitCidExists(tCtx, bs[0].Cid()))
		}()
		go func() {
			defer wg.Done()
			require.NoError(t, fx.WaitCidExists(tCtx, bs[1].Cid()))
		}()
		time.Sleep(time.Second / 2)
		fx.OnBlockUploaded(ctx, bs[0], bs[1])
		wg.Wait()
	})

	t.Run("ctx done", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.Finish(t)
		bs := testutil.NewRandBlocks(1)
		tCtx, cancel := context.WithTimeout(ctx, time.Second/10)
		defer cancel()

		assert.ErrorIs(t, fx.WaitCidExists(tCtx, bs[0].Cid()), context.DeadlineExceeded)
	})
}
