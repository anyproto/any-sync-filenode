package index

import (
	"context"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const cidsChannel = "cidsChan"

func (ri *redisIndex) WaitCidExists(ctx context.Context, k cid.Cid) (err error) {
	ck := cidKey(k)
	exists, release, err := ri.acquireKey(ctx, ck)
	if err != nil {
		return err
	}
	if exists {
		release()
		return nil
	}

	ri.cidSubscriptionsMu.Lock()
	release()
	var ch = make(chan struct{})
	m := ri.cidSubscriptions[ck]
	if m == nil {
		m = make(map[chan struct{}]struct{})
	}
	m[ch] = struct{}{}
	ri.cidSubscriptions[ck] = m
	ri.cidSubscriptionsMu.Unlock()

	cleanup := func() {
		ri.cidSubscriptionsMu.Lock()
		m := ri.cidSubscriptions[ck]
		if m != nil {
			delete(m, ch)
			if len(m) == 0 {
				delete(ri.cidSubscriptions, ck)
			}
		}
		ri.cidSubscriptionsMu.Unlock()
	}

	select {
	case <-ctx.Done():
		cleanup()
		return ctx.Err()
	case <-ch:
		return
	}
}

func (ri *redisIndex) OnBlockUploaded(ctx context.Context, bs ...blocks.Block) {
	if _, err := ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, b := range bs {
			_ = pipe.Publish(ctx, cidsChannel, cidKey(b.Cid()))
		}
		return nil
	}); err != nil {
		log.WarnCtx(ctx, "can't publish uploaded cids", zap.Error(err))
	}
}

func (ri *redisIndex) subscription(ctx context.Context) {
	sub := ri.cl.Subscribe(ctx, cidsChannel)
	defer func() {
		_ = sub.Unsubscribe(ctx, cidsChannel)
	}()
	ch := sub.Channel()
	for msg := range ch {
		ri.handleSubscriptionMessage(msg.Payload)
	}
}

func (ri *redisIndex) handleSubscriptionMessage(msg string) {
	ri.cidSubscriptionsMu.Lock()
	defer ri.cidSubscriptionsMu.Unlock()

	chans := ri.cidSubscriptions[msg]
	if chans != nil {
		for ch := range chans {
			close(ch)
		}
		delete(ri.cidSubscriptions, msg)
	}
}
