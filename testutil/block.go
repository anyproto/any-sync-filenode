package testutil

import (
	"github.com/anyproto/any-sync/util/cidutil"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"io"
	"math/rand"
	"time"
)

func NewRandSpaceId() string {
	b := NewRandBlock(256)
	return b.Cid().String() + ".123456"
}

func NewRandBlock(size int) blocks.Block {
	var p = make([]byte, size)
	_, err := io.ReadFull(rand.New(rand.NewSource(time.Now().UnixNano())), p)
	if err != nil {
		panic("can't fill testdata from rand")
	}
	c, _ := cidutil.NewCidFromBytes(p)
	b, _ := blocks.NewBlockWithCid(p, cid.MustParse(c))
	return b
}

func BlocksToKeys(bs []blocks.Block) (cids []cid.Cid) {
	cids = make([]cid.Cid, len(bs))
	for i, b := range bs {
		cids[i] = b.Cid()
	}
	return
}

func NewRandBlocks(l int) []blocks.Block {
	var bs = make([]blocks.Block, l)
	for i := range bs {
		bs[i] = NewRandBlock(10 * l)
	}
	return bs
}

func NewRandCid() cid.Cid {
	return NewRandBlock(1024).Cid()
}
