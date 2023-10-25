package filedevstore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/commonfile/fileblockstore"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"

	"github.com/anyproto/any-sync-filenode/config"
	"github.com/anyproto/any-sync-filenode/store"
)

const CName = fileblockstore.CName

var log = logger.NewNamed(CName)

func New() store.Store {
	return &fsstore{}
}

type configSource interface {
	GetDevStore() config.FileDevStore
}

type fsstore struct {
	path string
}

func (s *fsstore) Init(a *app.App) (err error) {
	s.path = a.MustComponent("config").(configSource).GetDevStore().Path
	if s.path == "" {
		return fmt.Errorf("you must specify path for local fsstore")
	}
	return
}

func (s *fsstore) Name() (name string) {
	return CName
}

func (s *fsstore) Run(ctx context.Context) (err error) {
	return
}

func (s *fsstore) Get(ctx context.Context, k cid.Cid) (blocks.Block, error) {
	val, err := os.ReadFile(filepath.Join(s.path, k.String()))
	if err != nil {
		return nil, err
	}
	return blocks.NewBlockWithCid(val, k)
}

func (s *fsstore) GetMany(ctx context.Context, ks []cid.Cid) <-chan blocks.Block {
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

func (s *fsstore) Add(ctx context.Context, bs []blocks.Block) error {
	for _, b := range bs {
		if err := os.WriteFile(filepath.Join(s.path, b.Cid().String()), b.RawData(), 0777); err != nil {
			return err
		}
	}
	return nil
}

func (s *fsstore) Delete(ctx context.Context, c cid.Cid) error {
	return os.Remove(filepath.Join(s.path, c.String()))
}

func (s *fsstore) Close(ctx context.Context) (err error) {
	return
}

func (s *fsstore) DeleteMany(ctx context.Context, toDelete []cid.Cid) error {
	for _, k := range toDelete {
		_ = s.Delete(ctx, k)
	}
	return nil
}

func (s *fsstore) IndexGet(ctx context.Context, key string) (value []byte, err error) {
	return os.ReadFile(filepath.Join(s.path, key))
}

func (s *fsstore) IndexPut(ctx context.Context, key string, value []byte) (err error) {
	return os.WriteFile(filepath.Join(s.path, key), value, 0777)
}
