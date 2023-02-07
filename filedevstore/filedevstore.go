package filedevstore

import (
	"context"
	"fmt"
	"github.com/anytypeio/any-sync-filenode/config"
	"github.com/anytypeio/any-sync-filenode/serverstore"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/app/logger"
	"github.com/anytypeio/any-sync/commonfile/fileblockstore"
	"github.com/anytypeio/any-sync/commonfile/fileproto"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-libipfs/blocks"
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
	serverstore.ServerStore
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

func (s *store) Check(ctx context.Context, spaceId string, cids ...cid.Cid) (result []*fileproto.BlockAvailability, err error) {
	for _, c := range cids {
		filename := filepath.Join(s.path, c.String())
		_, statErr := os.Stat(filename)
		var status = fileproto.AvailabilityStatus_Exists
		if statErr != nil {
			if os.IsNotExist(statErr) {
				status = fileproto.AvailabilityStatus_NotExists
			} else {
				return nil, statErr
			}
		}
		result = append(result, &fileproto.BlockAvailability{
			Cid:    c.Bytes(),
			Status: status,
		})
	}
	return
}
