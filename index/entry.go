package index

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/anyproto/any-sync-filenode/index/indexproto"
)

type fileEntry struct {
	*indexproto.FileEntry
}

func (f *fileEntry) Exists(c string) (ok bool) {
	return slices.Contains(f.Cids, c)
}

func (f *fileEntry) Save(ctx context.Context, k Key, fileId string, cl redis.Pipeliner) {
	f.UpdateTime = time.Now().Unix()
	data, err := f.Marshal()
	if err != nil {
		return
	}
	cl.HSet(ctx, spaceKey(k), fileKey(fileId), data)
}

func (ri *redisIndex) getFileEntry(ctx context.Context, k Key, fileId string) (entry *fileEntry, isCreated bool, err error) {
	result, err := ri.cl.HGet(ctx, spaceKey(k), fileKey(fileId)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return
	}
	if errors.Is(err, redis.Nil) {
		return &fileEntry{
			FileEntry: &indexproto.FileEntry{
				CreateTime: time.Now().Unix(),
			},
		}, true, nil
	}
	fileEntryProto := &indexproto.FileEntry{}
	if err = fileEntryProto.Unmarshal([]byte(result)); err != nil {
		return
	}
	return &fileEntry{FileEntry: fileEntryProto}, false, nil
}

type spaceEntry struct {
	Id string
	*indexproto.SpaceEntry
}

func (f *spaceEntry) Save(ctx context.Context, k Key, cl redis.Pipeliner) {
	f.UpdateTime = time.Now().Unix()
	data, err := f.Marshal()
	if err != nil {
		return
	}
	cl.HSet(ctx, spaceKey(k), infoKey, data)
}

func (ri *redisIndex) getSpaceEntry(ctx context.Context, key Key) (entry *spaceEntry, err error) {
	result, err := ri.cl.HGet(ctx, spaceKey(key), infoKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return
	}
	if errors.Is(err, redis.Nil) {
		now := time.Now().Unix()
		return &spaceEntry{
			Id: key.SpaceId,
			SpaceEntry: &indexproto.SpaceEntry{
				CreateTime: now,
				UpdateTime: now,
				GroupId:    key.GroupId,
			},
		}, nil
	}
	spaceEntryProto := &indexproto.SpaceEntry{}
	if err = spaceEntryProto.Unmarshal([]byte(result)); err != nil {
		return
	}
	return &spaceEntry{SpaceEntry: spaceEntryProto, Id: key.SpaceId}, nil
}

type groupEntry struct {
	*indexproto.GroupEntry
}

func (f *groupEntry) Save(ctx context.Context, cl redis.Cmdable) {
	f.UpdateTime = time.Now().Unix()
	data, err := f.Marshal()
	if err != nil {
		return
	}
	cl.HSet(ctx, groupKey(Key{GroupId: f.GroupId}), infoKey, data)
}

func (f *groupEntry) AddSpaceId(spaceId string) {
	if !slices.Contains(f.SpaceIds, spaceId) {
		f.SpaceIds = append(f.SpaceIds, spaceId)
	}
}

func (ri *redisIndex) getGroupEntry(ctx context.Context, key Key) (entry *groupEntry, err error) {
	result, err := ri.cl.HGet(ctx, groupKey(key), infoKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return
	}
	if errors.Is(err, redis.Nil) {
		now := time.Now().Unix()
		return &groupEntry{
			GroupEntry: &indexproto.GroupEntry{
				GroupId:      key.GroupId,
				CreateTime:   now,
				UpdateTime:   now,
				Size_:        0,
				Limit:        ri.defaultLimit,
				AccountLimit: ri.defaultLimit,
			},
		}, nil
	}
	groupEntryProto := &indexproto.GroupEntry{}
	if err = groupEntryProto.Unmarshal([]byte(result)); err != nil {
		return
	}
	groupEntryProto.GroupId = key.GroupId
	if groupEntryProto.AccountLimit == 0 {
		groupEntryProto.Limit = ri.defaultLimit
		groupEntryProto.AccountLimit = ri.defaultLimit
	}
	return &groupEntry{GroupEntry: groupEntryProto}, nil
}

type groupSpaceEntry struct {
	space       *spaceEntry
	group       *groupEntry
	spaceExists bool
	groupExists bool
}
