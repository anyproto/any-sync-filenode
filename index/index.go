//go:generate mockgen -destination mock_index/mock_index.go github.com/anyproto/any-sync-filenode/index Index
package index

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/util/periodicsync"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/redis/go-redis/v9"

	"github.com/anyproto/any-sync-filenode/config"
	"github.com/anyproto/any-sync-filenode/redisprovider"
	"github.com/anyproto/any-sync-filenode/store/s3store"
)

const CName = "filenode.index"

var log = logger.NewNamed(CName)

var (
	ErrCidsNotExist = errors.New("cids not exist")
)

type Index interface {
	FileBind(ctx context.Context, key Key, fileId string, cidEntries *CidEntries) (err error)
	FileUnbind(ctx context.Context, kye Key, fileIds ...string) (err error)
	FileInfo(ctx context.Context, key Key, fileIds ...string) (fileInfo []FileInfo, err error)
	FilesList(ctx context.Context, key Key) (fileIds []string, err error)

	GroupInfo(ctx context.Context, groupId string) (info GroupInfo, err error)
	SpaceInfo(ctx context.Context, key Key) (info SpaceInfo, err error)

	BlocksGetNonExistent(ctx context.Context, bs []blocks.Block) (nonExistent []blocks.Block, err error)
	BlocksLock(ctx context.Context, bs []blocks.Block) (unlock func(), err error)
	BlocksAdd(ctx context.Context, bs []blocks.Block) (err error)
	OnBlockUploaded(ctx context.Context, bs ...blocks.Block)

	WaitCidExists(ctx context.Context, c cid.Cid) (err error)
	CidExists(ctx context.Context, c cid.Cid) (ok bool, err error)
	CidEntries(ctx context.Context, cids []cid.Cid) (entries *CidEntries, err error)
	CidEntriesByBlocks(ctx context.Context, bs []blocks.Block) (entries *CidEntries, err error)
	CidExistsInSpace(ctx context.Context, key Key, cids []cid.Cid) (exists []cid.Cid, err error)

	SetGroupLimit(ctx context.Context, groupId string, limit uint64) (err error)
	SetSpaceLimit(ctx context.Context, key Key, limit uint64) (err error)
	CheckLimits(ctx context.Context, key Key) error

	Migrate(ctx context.Context, key Key) error

	SpaceDelete(ctx context.Context, key Key) (ok bool, err error)
	app.ComponentRunnable
}

func New() Index {
	return &redisIndex{}
}

type Key struct {
	GroupId string
	SpaceId string
}

type GroupInfo struct {
	BytesUsage   uint64
	CidsCount    uint64
	AccountLimit uint64
	Limit        uint64
	SpaceIds     []string
}

type SpaceInfo struct {
	BytesUsage uint64
	CidsCount  uint64
	Limit      uint64
	FileCount  uint32
}

type FileInfo struct {
	BytesUsage uint64
	CidsCount  uint64
}

/*
	Redis db structure:
		CIDS:
			c:{cid}: proto(Entry)
			cidCount.{system}: int
			cidSizeSum.{system}: int
		STORES:
			g:{groupId}: map
				c:{cidId} -> int(refCount)
				info: proto(GroupEntry}
			s:{spaceId}: map
				f:{fileId}: proto(FileEntry)
				c:{cidId} -> int(refCount)
				info: proto(SpaceEntry)

*/

type redisIndex struct {
	cl           redis.UniversalClient
	redsync      *redsync.Redsync
	persistStore persistentStore
	persistTtl   time.Duration
	ticker       periodicsync.PeriodicSync
	defaultLimit uint64

	cidSubscriptionsMu sync.Mutex
	cidSubscriptions   map[string]map[chan struct{}]struct{}

	ctx       context.Context
	ctxCancel context.CancelFunc
}

func (ri *redisIndex) Init(a *app.App) (err error) {
	ri.cl = a.MustComponent(redisprovider.CName).(redisprovider.RedisProvider).Redis()
	ri.persistStore = a.MustComponent(s3store.CName).(persistentStore)
	ri.redsync = redsync.New(goredis.NewPool(ri.cl))
	conf := app.MustComponent[*config.Config](a)

	ri.persistTtl = time.Second * time.Duration(conf.PersistTtl)
	if ri.persistTtl == 0 {
		ri.persistTtl = time.Hour
	}
	ri.defaultLimit = conf.DefaultLimit
	if ri.defaultLimit == 0 {
		ri.defaultLimit = 1 << 30
	}
	ri.cidSubscriptions = make(map[string]map[chan struct{}]struct{})
	ri.ctx, ri.ctxCancel = context.WithCancel(context.Background())
	return
}

func (ri *redisIndex) Name() (name string) {
	return CName
}

func (ri *redisIndex) Run(ctx context.Context) (err error) {
	ri.ticker = periodicsync.NewPeriodicSync(60, time.Minute*10, func(ctx context.Context) error {
		ri.PersistKeys(ctx)
		return nil
	}, log)
	ri.ticker.Run()
	go ri.subscription(ctx)
	return
}

func (ri *redisIndex) FileInfo(ctx context.Context, key Key, fileIds ...string) (fileInfos []FileInfo, err error) {
	_, release, err := ri.AcquireKey(ctx, spaceKey(key))
	if err != nil {
		return
	}
	defer release()
	fileInfos = make([]FileInfo, len(fileIds))
	for i, fileId := range fileIds {
		fEntry, _, err := ri.getFileEntry(ctx, key, fileId)
		if err != nil {
			return nil, err
		}
		fileInfos[i] = FileInfo{
			BytesUsage: fEntry.Size_,
			CidsCount:  uint64(len(fEntry.Cids)),
		}
	}
	return
}

func (ri *redisIndex) FilesList(ctx context.Context, key Key) (fileIds []string, err error) {
	sk := spaceKey(key)
	_, release, err := ri.AcquireKey(ctx, sk)
	if err != nil {
		return
	}
	defer release()
	allKeys, err := ri.cl.HKeys(ctx, sk).Result()
	if err != nil {
		return
	}
	for _, k := range allKeys {
		if strings.HasPrefix(k, "f:") {
			fileIds = append(fileIds, k[2:])
		}
	}
	return
}

func (ri *redisIndex) BlocksGetNonExistent(ctx context.Context, bs []blocks.Block) (nonExistent []blocks.Block, err error) {
	for _, b := range bs {
		ex, err := ri.CheckKey(ctx, cidKey(b.Cid()))
		if err != nil {
			return nil, err
		}
		if !ex {
			nonExistent = append(nonExistent, b)
		}
	}
	return
}

func (ri *redisIndex) BlocksLock(ctx context.Context, bs []blocks.Block) (unlock func(), err error) {
	var lockers = make([]*redsync.Mutex, 0, len(bs))

	unlock = func() {
		for _, l := range lockers {
			_, _ = l.Unlock()
		}
	}

	for _, b := range bs {
		l := ri.redsync.NewMutex("_lock:b:"+b.Cid().String(), redsync.WithExpiry(time.Minute))
		if err = l.LockContext(ctx); err != nil {
			unlock()
			return nil, err
		}
		lockers = append(lockers, l)
	}
	return
}

func (ri *redisIndex) GroupInfo(ctx context.Context, groupId string) (info GroupInfo, err error) {
	_, release, err := ri.AcquireKey(ctx, groupKey(Key{GroupId: groupId}))
	if err != nil {
		return
	}
	defer release()
	sEntry, err := ri.getGroupEntry(ctx, Key{GroupId: groupId})
	if err != nil {
		return
	}
	return GroupInfo{
		BytesUsage:   sEntry.Size_,
		CidsCount:    sEntry.CidCount,
		AccountLimit: sEntry.AccountLimit,
		Limit:        sEntry.Limit,
		SpaceIds:     sEntry.SpaceIds,
	}, nil
}

func (ri *redisIndex) SpaceInfo(ctx context.Context, key Key) (info SpaceInfo, err error) {
	_, release, err := ri.AcquireKey(ctx, spaceKey(key))
	if err != nil {
		return
	}
	defer release()
	sEntry, err := ri.getSpaceEntry(ctx, key)
	if err != nil {
		return
	}
	return SpaceInfo{
		BytesUsage: sEntry.Size_,
		CidsCount:  sEntry.CidCount,
		Limit:      sEntry.Limit,
		FileCount:  sEntry.FileCount,
	}, nil
}

func (ri *redisIndex) Close(ctx context.Context) error {
	if ri.ticker != nil {
		ri.ticker.Close()
	}
	if ri.ctxCancel != nil {
		ri.ctxCancel()
	}
	return nil
}

func cidKey(c cid.Cid) string {
	return "c:" + c.String()
}

func spaceKey(k Key) string {
	hash := strconv.FormatUint(uint64(xxhash.ChecksumString32(k.GroupId)), 36)
	return "s:" + k.SpaceId + ".{" + hash + "}"
}

func groupKey(k Key) string {
	hash := strconv.FormatUint(uint64(xxhash.ChecksumString32(k.GroupId)), 36)
	return "g:" + k.GroupId + ".{" + hash + "}"
}

func fileKey(fileId string) string {
	return "f:" + fileId
}

const infoKey = "info"
