//go:generate mockgen -destination mock_serverstore/mock_serverstore.go github.com/anytypeio/any-sync-filenode/serverstore Service,ServerStore
package serverstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/anytypeio/any-sync-filenode/index"
	"github.com/anytypeio/any-sync-filenode/index/redisindex"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/app/logger"
	"github.com/anytypeio/any-sync/commonfile/fileblockstore"
	"github.com/anytypeio/any-sync/commonfile/fileproto"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-libipfs/blocks"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

const CName = "filenode.serverstore"

var log = logger.NewNamed(CName)

var closedResult = make(chan blocks.Block)

var (
	ErrCidsNotExists = errors.New("cids not exists")
)

func init() {
	close(closedResult)
}

func New() Service {
	return new(serverStore)
}

type ServerStore interface {
	fileblockstore.BlockStore
	DeleteMany(ctx context.Context, toDelete []cid.Cid) error
	app.Component
}
type Service interface {
	fileblockstore.BlockStore
	Check(ctx context.Context, spaceId string, cids ...cid.Cid) (result []*fileproto.BlockAvailability, err error)
	BlocksBind(ctx context.Context, spaceId string, cids ...cid.Cid) (err error)
	app.Component
}

type serverStore struct {
	index index.Index
	store ServerStore
}

func (s *serverStore) Init(a *app.App) (err error) {
	s.store = a.MustComponent(fileblockstore.CName).(ServerStore)
	s.index = a.MustComponent(redisindex.CName).(index.Index)
	return nil
}

func (s *serverStore) Name() (name string) {
	return CName
}

func (s *serverStore) Get(ctx context.Context, k cid.Cid) (blocks.Block, error) {
	exists, err := s.index.Exists(ctx, k)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fileblockstore.ErrCIDNotFound
	}
	return s.store.Get(ctx, k)
}

func (s *serverStore) GetMany(ctx context.Context, ks []cid.Cid) <-chan blocks.Block {
	var err error
	ks, err = s.index.FilterExistingOnly(ctx, ks)
	if err != nil {
		log.Warn("filterNonExists error", zap.Error(err))
		return closedResult
	}
	return s.store.GetMany(ctx, ks)
}

func (s *serverStore) Add(ctx context.Context, bs []blocks.Block) error {
	spaceId := fileblockstore.CtxGetSpaceId(ctx)
	if err := s.validateSpaceId(ctx, spaceId); err != nil {
		return err
	}
	toUpload, err := s.index.GetNonExistentBlocks(ctx, bs)
	if err != nil {
		return err
	}
	if len(toUpload) > 0 {
		if err = s.store.Add(ctx, toUpload); err != nil {
			return err
		}
	}
	return s.index.Bind(ctx, spaceId, bs)
}

func (s *serverStore) Delete(ctx context.Context, k cid.Cid) error {
	return s.DeleteMany(ctx, []cid.Cid{k})
}

func (s *serverStore) DeleteMany(ctx context.Context, ks []cid.Cid) error {
	spaceId := fileblockstore.CtxGetSpaceId(ctx)
	if err := s.validateSpaceId(ctx, spaceId); err != nil {
		return err
	}
	toDelete, err := s.index.UnBind(ctx, spaceId, ks)
	if err != nil {
		return err
	}
	if len(toDelete) > 0 {
		return s.store.DeleteMany(ctx, toDelete)
	}
	return nil
}

func (s *serverStore) Check(ctx context.Context, spaceId string, cids ...cid.Cid) (result []*fileproto.BlockAvailability, err error) {
	result = make([]*fileproto.BlockAvailability, 0, len(cids))
	inSpace, err := s.index.ExistsInSpace(ctx, spaceId, cids)
	if err != nil {
		return nil, err
	}
	var inSpaceM = make(map[string]struct{})
	for _, inSp := range inSpace {
		inSpaceM[inSp.KeyString()] = struct{}{}
	}

	for _, k := range cids {
		var res = &fileproto.BlockAvailability{
			Cid:    k.Bytes(),
			Status: fileproto.AvailabilityStatus_NotExists,
		}
		if _, ok := inSpaceM[k.KeyString()]; ok {
			res.Status = fileproto.AvailabilityStatus_ExistsInSpace
		} else {
			var ex bool
			if ex, err = s.index.Exists(ctx, k); err != nil {
				return nil, err
			} else if ex {
				res.Status = fileproto.AvailabilityStatus_Exists
			}
		}
		result = append(result, res)
	}
	return
}

func (s *serverStore) BlocksBind(ctx context.Context, spaceId string, cids ...cid.Cid) (err error) {
	if err = s.validateSpaceId(ctx, spaceId); err != nil {
		return err
	}
	// check maybe it's already bound
	existingInSpace, err := s.index.ExistsInSpace(ctx, spaceId, cids)
	if err != nil {
		return
	}
	if len(existingInSpace) == len(cids) {
		// all cids bound
		return nil
	}

	// check that we have all cids uploaded
	existing, err := s.index.FilterExistingOnly(ctx, cids)
	if err != nil {
		return err
	}
	if len(existing) != len(cids) {
		return ErrCidsNotExists
	}

	var cidsToBound []cid.Cid
	if len(existingInSpace) > 0 {
		// if we have some cids bound, filter it
		cidsToBound = make([]cid.Cid, 0, len(cids)-len(existingInSpace))
		for _, c := range cids {
			if !slices.Contains(existingInSpace, c) {
				cidsToBound = append(cidsToBound, c)
			}
		}
	} else {
		cidsToBound = cids
	}

	var bs = make([]blocks.Block, 0, len(cids))
	for b := range s.store.GetMany(ctx, cids) {
		bs = append(bs, b)
	}
	if len(bs) != len(cids) {
		return ErrCidsNotExists
	}
	return s.index.Bind(ctx, spaceId, bs)
}

func (s *serverStore) validateSpaceId(ctx context.Context, spaceId string) (err error) {
	if spaceId == "" {
		return fmt.Errorf("empty space id")
	}
	// TODO: make validation checks here
	return
}
