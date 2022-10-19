package storage

import (
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/commonspace/spacesyncproto"
	spacestorage "github.com/anytypeio/go-anytype-infrastructure-experiments/common/commonspace/storage"
	storage2 "github.com/anytypeio/go-anytype-infrastructure-experiments/common/pkg/acl/storage"
	"github.com/dgraph-io/badger/v3"
	"sync"
)

type spaceStorage struct {
	spaceId    string
	objDb      *badger.DB
	keys       spaceKeys
	aclStorage storage2.ListStorage
	header     *spacesyncproto.RawSpaceHeaderWithId
	mx         sync.Mutex
}

func newSpaceStorage(objDb *badger.DB, spaceId string) (store spacestorage.SpaceStorage, err error) {
	keys := newSpaceKeys(spaceId)
	err = objDb.View(func(txn *badger.Txn) error {
		header, err := getTxn(txn, keys.HeaderKey())
		if err != nil {
			return err
		}

		aclStorage, err := newListStorage(spaceId, objDb, txn)
		if err != nil {
			return err
		}

		store = &spaceStorage{
			spaceId: spaceId,
			objDb:   objDb,
			keys:    keys,
			header: &spacesyncproto.RawSpaceHeaderWithId{
				RawHeader: header,
				Id:        spaceId,
			},
			aclStorage: aclStorage,
		}
		return nil
	})
	if err == badger.ErrKeyNotFound {
		err = spacesyncproto.ErrSpaceMissing
	}
	return
}

func createSpaceStorage(db *badger.DB, payload spacestorage.SpaceStorageCreatePayload) (store spacestorage.SpaceStorage, err error) {
	keys := newSpaceKeys(payload.SpaceHeaderWithId.Id)
	if hasDB(db, keys.HeaderKey()) {
		err = spacesyncproto.ErrSpaceExists
		return
	}
	err = db.Update(func(txn *badger.Txn) error {
		aclStorage, err := createListStorage(payload.SpaceHeaderWithId.Id, db, txn, payload.RecWithId)
		if err != nil {
			return err
		}

		err = txn.Set(keys.HeaderKey(), payload.SpaceHeaderWithId.RawHeader)
		if err != nil {
			return err
		}

		store = &spaceStorage{
			spaceId:    payload.SpaceHeaderWithId.Id,
			objDb:      db,
			keys:       keys,
			aclStorage: aclStorage,
			header:     payload.SpaceHeaderWithId,
		}
		return nil
	})
	return
}

func (s *spaceStorage) ID() (string, error) {
	return s.spaceId, nil
}

func (s *spaceStorage) TreeStorage(id string) (storage2.TreeStorage, error) {
	return newTreeStorage(s.objDb, s.spaceId, id)
}

func (s *spaceStorage) CreateTreeStorage(payload storage2.TreeStorageCreatePayload) (ts storage2.TreeStorage, err error) {
	// we have mutex here, so we prevent overwriting the heads of a tree on concurrent creation
	s.mx.Lock()
	defer s.mx.Unlock()

	return createTreeStorage(s.objDb, s.spaceId, payload)
}

func (s *spaceStorage) ACLStorage() (storage2.ListStorage, error) {
	return s.aclStorage, nil
}

func (s *spaceStorage) SpaceHeader() (header *spacesyncproto.RawSpaceHeaderWithId, err error) {
	return s.header, nil
}

func (s *spaceStorage) StoredIds() (ids []string, err error) {
	err = s.objDb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = s.keys.TreeRootPrefix()

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			id := item.Key()
			if len(id) <= len(s.keys.TreeRootPrefix())+1 {
				continue
			}
			id = id[len(s.keys.TreeRootPrefix())+1:]
			ids = append(ids, string(id))
		}
		return nil
	})
	return
}

func (s *spaceStorage) Close() (err error) {
	return nil
}