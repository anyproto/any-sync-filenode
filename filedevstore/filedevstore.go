package filedevstore

import (
	"context"
	"fmt"
	"github.com/anytypeio/any-sync-filenode/config"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/app/logger"
	"github.com/anytypeio/any-sync/commonfile/fileblockstore"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"os"
	"path/filepath"
)

const CName = fileblockstore.CName

var log = logger.NewNamed(CName)

func New() Store {
	return &store{}
}

type Store interface {
	app.ComponentRunnable
	fileblockstore.BlockStore
}

type configSource interface {
	GetDevStore() config.FileDevStore
}

type store struct {
	path string
}

func (s *store) Init(a *app.App) (err error) {
	s.path = a.MustComponent("config").(configSource).GetDevStore().Path
	if s.path == "" {
		return fmt.Errorf("you must specify path for local store")
	}
	return
}

func (s *store) Name() (name string) {
	return CName
}

func (s *store) Run(ctx context.Context) (err error) {
	return
}

func (s *store) Get(ctx context.Context, k cid.Cid) (blocks.Block, error) {
	val, err := os.ReadFile(filepath.Join(s.path, k.String()))
	if err != nil {
		return nil, err
	}
	return blocks.NewBlockWithCid(val, k)
}

func (s *store) GetMany(ctx context.Context, ks []cid.Cid) <-chan blocks.Block {
	var res = make(chan blocks.Block)
	go func() {
		defer close(res)
		for _, k := range ks {
			b, err := s.Get(ctx, k)
			if err != nil {
				continue
			}
			select {
			case <-ctx.Done():
				return
			case res <- b:
			}
		}
	}()
	return res
}

func (s *store) Add(ctx context.Context, bs []blocks.Block) error {
	for _, b := range bs {
		if err := os.WriteFile(filepath.Join(s.path, b.Cid().String()), b.RawData(), 0777); err != nil {
			return err
		}
	}
	return nil
}

func (s *store) Delete(ctx context.Context, c cid.Cid) error {
	return os.Remove(filepath.Join(s.path, c.String()))
}

func (s *store) Close(ctx context.Context) (err error) {
	return
}
