package index

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/anyproto/any-sync-filenode/index/indexproto"
)

func (ri *redisIndex) CheckAndMoveOwnership(ctx context.Context, key Key, oldIdentity string, aclRecordIndex int) (err error) {
	oKey := OwnerKey(key.SpaceId)
	_, release, err := ri.AcquireKey(ctx, oKey)
	if err != nil {
		return
	}
	defer release()

	var (
		ownerData = &indexproto.OwnershipRecord{
			OwnerId: oldIdentity,
		}
		notExists bool
	)
	result, err := ri.cl.Get(ctx, oKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			notExists = true
			err = nil
		} else {
			return
		}
	}

	if !notExists {
		if err = ownerData.UnmarshalVT([]byte(result)); err != nil {
			return
		}
	} else if oldIdentity == "" {
		return fmt.Errorf("previous owner not found")
	}

	var needUpdate bool
	if ownerData.OwnerId != key.GroupId && ownerData.AclRecordIndex <= int64(aclRecordIndex) {
		if err = ri.Move(ctx, key, Key{SpaceId: key.SpaceId, GroupId: ownerData.OwnerId}); err != nil {
			return
		}
		ownerData.OwnerId = key.GroupId
		needUpdate = true
	}

	if ownerData.AclRecordIndex < int64(aclRecordIndex) {
		ownerData.AclRecordIndex = int64(aclRecordIndex)
		needUpdate = true
	}

	if !needUpdate {
		return
	}

	data, err := ownerData.MarshalVT()
	if err != nil {
		return err
	}
	return ri.cl.Set(ctx, oKey, data, 0).Err()
}

// moveDestScript atomically applies the destination side of a space move:
// restores the dumped space hash at the new key, writes the updated space and
// group info entries and increments the group cid refcounts.
// A marker field in the dest group hash makes the script a no-op when this
// move was already applied but the src cleanup didn't finish (crash between
// the two scripts); the marker is removed after the src side succeeds.
// KEYS: [1] dest space key, [2] dest group key
// ARGV: [1] marker field, [2] marker token (src group id), [3] space hash dump
// (empty - skip restore), [4] space info (empty - keep existing), [5] group info,
// [6..] cid field/delta pairs
var moveDestScript = redis.NewScript(`
if redis.call('HGET', KEYS[2], ARGV[1]) == ARGV[2] then
	return 0
end
if ARGV[3] ~= '' then
	redis.call('RESTORE', KEYS[1], 0, ARGV[3], 'REPLACE')
end
if ARGV[4] ~= '' then
	redis.call('HSET', KEYS[1], 'info', ARGV[4])
end
redis.call('HSET', KEYS[2], 'info', ARGV[5])
for i = 6, #ARGV, 2 do
	redis.call('HINCRBY', KEYS[2], ARGV[i], ARGV[i + 1])
end
redis.call('HSET', KEYS[2], ARGV[1], ARGV[2])
return 1
`)

// moveSrcScript atomically applies the source side of a space move:
// decrements the group cid refcounts (dropping the emptied fields), writes the
// updated group info entry and deletes the moved space hash.
// KEYS: [1] src space key, [2] src group key
// ARGV: [1] group info, [2] '1' to delete the space key, [3..] cid field/delta pairs
var moveSrcScript = redis.NewScript(`
for i = 3, #ARGV, 2 do
	local v = redis.call('HINCRBY', KEYS[2], ARGV[i], -ARGV[i + 1])
	if v <= 0 then
		redis.call('HDEL', KEYS[2], ARGV[i])
	end
end
redis.call('HSET', KEYS[2], 'info', ARGV[1])
if ARGV[2] == '1' then
	redis.call('DEL', KEYS[1])
end
return 1
`)

// moveMarker names the field in the dest group hash guarding against
// re-applying the dest side of a crashed move; its value is the src group id
func moveMarker(spaceId string) string {
	return "mv:" + spaceId
}

// movePlan holds everything precomputed under the space/group locks so the
// move applies as two atomic scripts: one per cluster slot side.
type movePlan struct {
	sSK, dSK string
	sGK, dGK string
	// marker field in the dest group hash guarding against re-applying the
	// dest side when a crashed move is retried
	marker      string
	markerToken string
	// dump of the whole src space hash; empty when the src key is missing or
	// the move stays on the same key (group hash collision)
	dump string
	// updated space info entry; nil means keep the existing dest info
	// (re-run of an already applied move)
	spaceInfo     []byte
	srcGroupInfo  []byte
	destGroupInfo []byte
	// per-cid refcount transfer: space refcounts count file bindings, so the
	// group refcounts move by the same amount
	cidFields []string
	cidDeltas []int64
	cids      *CidEntries
}

func (ri *redisIndex) Move(ctx context.Context, dest, src Key) (err error) {
	log.InfoCtx(ctx, "move space", zap.String("spaceId", dest.SpaceId), zap.String("src", src.GroupId), zap.String("dest", dest.GroupId))
	if dest.SpaceId != src.SpaceId {
		return fmt.Errorf("spaceId should be the same for both keys")
	}
	// accept a space already pointing at the dest group: with colliding group
	// hashes the dest side of a crashed move rewrites the shared space key
	srcEntry, srcRelease, err := ri.acquireSpace(ctx, src, dest.GroupId)
	if err != nil {
		return
	}
	defer srcRelease()

	_, destGRelease, err := ri.AcquireKey(ctx, GroupKey(dest))
	if err != nil {
		return
	}
	defer destGRelease()

	destGroup, err := ri.getGroupEntry(ctx, dest)
	if err != nil {
		return
	}

	if srcEntry.space.GetGroupId() == dest.GroupId {
		// the dest side was already applied by a previous attempt; without the
		// marker the whole move completed - nothing to do
		token, e := ri.cl.HGet(ctx, GroupKey(dest), moveMarker(src.SpaceId)).Result()
		if e != nil && !errors.Is(e, redis.Nil) {
			return e
		}
		if token != src.GroupId {
			return nil
		}
	}

	// in case of hash collision, we don't need to lock the space key twice
	destSpaceExists := srcEntry.spaceExists
	if SpaceKey(src) != SpaceKey(dest) {
		var destSRelease func()
		destSpaceExists, destSRelease, err = ri.AcquireKey(ctx, SpaceKey(dest))
		if err != nil {
			return err
		}
		defer destSRelease()
	}

	plan, err := ri.prepareMove(ctx, dest, src, srcEntry, destGroup, destSpaceExists)
	if err != nil {
		return
	}
	defer plan.cids.Release()

	if err = ri.applyMoveDest(ctx, plan); err != nil {
		return
	}
	if err = ri.applyMoveSrc(ctx, plan); err != nil {
		return
	}
	return ri.cl.HDel(ctx, plan.dGK, plan.marker).Err()
}

func (ri *redisIndex) prepareMove(ctx context.Context, dest, src Key, srcEntry groupSpaceEntry, destGroup *groupEntry, destSpaceExists bool) (plan *movePlan, err error) {
	plan = &movePlan{
		sSK:         SpaceKey(src),
		dSK:         SpaceKey(dest),
		sGK:         GroupKey(src),
		dGK:         GroupKey(dest),
		marker:      moveMarker(src.SpaceId),
		markerToken: src.GroupId,
	}

	// collect the space cid fields
	keys, err := ri.cl.HKeys(ctx, plan.sSK).Result()
	if err != nil {
		return nil, err
	}
	var cidFields = make([]string, 0, len(keys))
	for _, k := range keys {
		if strings.HasPrefix(k, "c:") {
			cidFields = append(cidFields, k)
		}
	}

	// fetch the refcounts and the space hash dump in one pipeline;
	// isolated spaces don't count against the group refcounts
	var (
		needRefs = srcEntry.space.Limit == 0 && len(cidFields) != 0
		needDump = plan.sSK != plan.dSK && srcEntry.spaceExists

		spaceRefsCmd, destRefsCmd, srcRefsCmd *redis.SliceCmd
		dumpCmd                               *redis.StringCmd
	)
	if needRefs || needDump {
		if _, err = ri.cl.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			if needRefs {
				spaceRefsCmd = pipe.HMGet(ctx, plan.sSK, cidFields...)
				destRefsCmd = pipe.HMGet(ctx, plan.dGK, cidFields...)
				srcRefsCmd = pipe.HMGet(ctx, plan.sGK, cidFields...)
			}
			if needDump {
				dumpCmd = pipe.Dump(ctx, plan.sSK)
			}
			return nil
		}); err != nil && !errors.Is(err, redis.Nil) {
			return nil, err
		}
		err = nil
	}
	if needDump {
		if plan.dump, err = dumpCmd.Result(); err != nil {
			if !errors.Is(err, redis.Nil) {
				return nil, err
			}
			plan.dump = ""
			err = nil
		}
	}

	// keep only the cids with a valid refcount to transfer
	var origIdx = make([]int, 0, len(cidFields))
	if needRefs {
		spaceRefs := spaceRefsCmd.Val()
		for i, k := range cidFields {
			ref, _ := spaceRefs[i].(string)
			delta, e := strconv.ParseInt(ref, 10, 64)
			if e != nil || delta <= 0 {
				log.WarnCtx(ctx, "move: skip cid with invalid space refcount",
					zap.String("spaceId", src.SpaceId), zap.String("field", k), zap.String("value", ref))
				continue
			}
			plan.cidFields = append(plan.cidFields, k)
			plan.cidDeltas = append(plan.cidDeltas, delta)
			origIdx = append(origIdx, i)
		}
	}

	// load the entries to lock the cids and get their sizes
	var cidStrings = make([]string, len(plan.cidFields))
	for i, k := range plan.cidFields {
		cidStrings[i] = k[2:]
	}
	if plan.cids, err = ri.CidEntriesByString(ctx, cidStrings); err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			plan.cids.Release()
		}
	}()

	// calculate the group entry sums
	if len(plan.cidFields) != 0 {
		destRefs := destRefsCmd.Val()
		srcRefs := srcRefsCmd.Val()
		for i, orig := range origIdx {
			size := plan.cids.entries[i].Size
			if destRefs[orig] == nil {
				destGroup.CidCount++
				destGroup.Size += size
			}
			srcRef, _ := srcRefs[orig].(string)
			oldVal, _ := strconv.ParseInt(srcRef, 10, 64)
			if oldVal > 0 && oldVal-plan.cidDeltas[i] <= 0 {
				if srcEntry.group.Size-size > srcEntry.group.Size {
					log.WarnCtx(ctx, "group: unable to decrement size", zap.Uint64("before", srcEntry.group.Size), zap.Uint64("size", size), zap.String("spaceId", src.SpaceId))
				} else {
					srcEntry.group.Size -= size
				}
				if srcEntry.group.CidCount != 0 {
					srcEntry.group.CidCount--
				} else {
					log.WarnCtx(ctx, "group: unable to decrement 0-ref", zap.String("spaceId", src.SpaceId))
				}
			}
		}
	}

	// an isolated space carries its reserved limit between the groups
	// (group.Limit = AccountLimit - sum of the isolated space limits);
	// the move can't be refused, so the dest side clamps instead of erroring -
	// CheckLimits enforces the account limit afterwards
	if srcEntry.space.Limit != 0 && src.GroupId != dest.GroupId {
		srcEntry.group.Limit += srcEntry.space.Limit
		if destGroup.Limit < srcEntry.space.Limit {
			log.WarnCtx(ctx, "move: dest group limit underflow",
				zap.Uint64("limit", destGroup.Limit), zap.Uint64("spaceLimit", srcEntry.space.Limit), zap.String("spaceId", src.SpaceId))
			destGroup.Limit = 0
		} else {
			destGroup.Limit -= srcEntry.space.Limit
		}
	}

	now := time.Now().Unix()

	// don't overwrite the dest space info when the src key is already gone:
	// it means a previous move was applied and the dest holds the actual entry
	if srcEntry.spaceExists || !destSpaceExists {
		srcEntry.space.GroupId = dest.GroupId
		srcEntry.space.UpdateTime = now
		if plan.spaceInfo, err = srcEntry.space.MarshalVT(); err != nil {
			return nil, err
		}
	}

	destGroup.AddSpaceId(src.SpaceId)
	destGroup.UpdateTime = now
	if plan.destGroupInfo, err = destGroup.MarshalVT(); err != nil {
		return nil, err
	}

	srcEntry.group.SpaceIds = slices.DeleteFunc(srcEntry.group.SpaceIds, func(spaceId string) bool {
		return spaceId == src.SpaceId
	})
	srcEntry.group.UpdateTime = now
	if plan.srcGroupInfo, err = srcEntry.group.MarshalVT(); err != nil {
		return nil, err
	}
	return plan, nil
}

func (ri *redisIndex) applyMoveDest(ctx context.Context, plan *movePlan) (err error) {
	args := make([]interface{}, 0, 5+2*len(plan.cidFields))
	var spaceInfo interface{} = ""
	if plan.spaceInfo != nil {
		spaceInfo = plan.spaceInfo
	}
	args = append(args, plan.marker, plan.markerToken, plan.dump, spaceInfo, plan.destGroupInfo)
	for i, k := range plan.cidFields {
		args = append(args, k, plan.cidDeltas[i])
	}
	return moveDestScript.Run(ctx, ri.cl, []string{plan.dSK, plan.dGK}, args...).Err()
}

func (ri *redisIndex) applyMoveSrc(ctx context.Context, plan *movePlan) (err error) {
	deleteKey := "0"
	if plan.sSK != plan.dSK {
		deleteKey = "1"
	}
	args := make([]interface{}, 0, 2+2*len(plan.cidFields))
	args = append(args, plan.srcGroupInfo, deleteKey)
	for i, k := range plan.cidFields {
		args = append(args, k, plan.cidDeltas[i])
	}
	return moveSrcScript.Run(ctx, ri.cl, []string{plan.sSK, plan.sGK}, args...).Err()
}
