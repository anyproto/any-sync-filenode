package filedevstore

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	anystore "github.com/anyproto/any-store"
	"github.com/anyproto/any-store/anyenc"
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

var (
	parserPool = &anyenc.ParserPool{}
	arenaPool  = &anyenc.ArenaPool{}
)

type configSource interface {
	GetDevStore() config.FileDevStore
}

type fsstore struct {
	path string
	db   anystore.DB
	data anystore.Collection
}

func (s *fsstore) Init(a *app.App) (err error) {
	s.path = a.MustComponent("config").(configSource).GetDevStore().Path
	if s.path == "" {
		return fmt.Errorf("you must specify path for local fsstore")
	}
	s.db, err = anystore.Open(context.Background(), filepath.Join(s.path, "data.db"), &anystore.Config{
		ReadConnections: 20,
		SQLiteConnectionOptions: map[string]string{
			"synchronous": "off",
		},
	})
	if err != nil {
		return err
	}
	s.data, err = s.db.Collection(context.Background(), "data")
	if err != nil {
		return err
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
	data, err := s.IndexGet(ctx, k.String())
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fileblockstore.ErrCIDNotFound
	}
	return blocks.NewBlockWithCid(data, k)
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
	tx, err := s.db.WriteTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	ctx = tx.Context()

	for _, b := range bs {
		if err = s.IndexPut(ctx, b.Cid().String(), b.RawData()); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *fsstore) Delete(ctx context.Context, c cid.Cid) error {
	return s.IndexDelete(ctx, c.String())
}

func (s *fsstore) Close(ctx context.Context) (err error) {
	if s.db != nil {
		err = s.db.Close()
	}
	return
}

func (s *fsstore) DeleteMany(ctx context.Context, toDelete []cid.Cid) error {
	for _, k := range toDelete {
		_ = s.Delete(ctx, k)
	}
	return nil
}

func (s *fsstore) IndexGet(ctx context.Context, key string) (value []byte, err error) {
	p := parserPool.Get()
	defer parserPool.Put(p)
	doc, err := s.data.FindIdWithParser(ctx, p, key)
	if err != nil {
		if errors.Is(err, anystore.ErrDocNotFound) {
			return nil, nil
		}
	}
	return slices.Clone(doc.Value().GetBytes("d")), nil
}

func (s *fsstore) IndexPut(ctx context.Context, key string, value []byte) (err error) {
	a := arenaPool.Get()
	defer arenaPool.Put(a)
	doc := a.NewObject()
	doc.Set("id", a.NewString(key))
	doc.Set("d", a.NewBinary(value))
	return s.data.UpsertOne(ctx, doc)
}

func (s *fsstore) IndexDelete(ctx context.Context, key string) (err error) {
	return s.data.DeleteId(ctx, key)
}
