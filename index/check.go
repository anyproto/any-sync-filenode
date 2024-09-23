package index

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/anyproto/any-sync-filenode/index/indexproto"
)

func (ri *redisIndex) Check(ctx context.Context, key Key, doFix bool) (checkResults []CheckResult, err error) {
	var toRelease []func()
	defer func() {
		for _, r := range toRelease {
			r()
		}
	}()

	// acquire locks and fetch entries
	gExists, gRelease, err := ri.AcquireKey(ctx, GroupKey(key))
	if err != nil {
		return
	}
	toRelease = append(toRelease, gRelease)
	if !gExists {
		return
	}
	gEntry, err := ri.getGroupEntry(ctx, key)
	if err != nil {
		return
	}

	var spaceChecks []*spaceContent

	// load spaces
	for _, spaceId := range gEntry.GetSpaceIds() {
		spaceKey := Key{GroupId: key.GroupId, SpaceId: spaceId}
		sExists, sRelease, aErr := ri.AcquireKey(ctx, SpaceKey(spaceKey))
		if aErr != nil {
			return nil, aErr
		}
		toRelease = append(toRelease, sRelease)
		if sExists {
			var (
				sEntry *spaceEntry
				check  *spaceContent
			)
			sEntry, err = ri.getSpaceEntry(ctx, spaceKey)
			if err != nil {
				return nil, err
			}
			check, err = ri.loadSpaceContent(ctx, spaceKey, sEntry)
			if err != nil {
				return nil, err
			}
			spaceChecks = append(spaceChecks, check)
		}
	}

	var sumRefs = make(map[string]uint64)
	var sumSize uint64
	for _, check := range spaceChecks {
		var results []CheckResult
		results, err = check.Check(ctx, ri)
		if err != nil {
			return nil, err
		}
		checkResults = append(checkResults, results...)
		for k, ref := range check.actualRefs {
			sumRefs[k] += ref
			sumSize += check.cidEntries[k].Size_
		}
	}

	groupCheck, err := ri.loadGroupContent(ctx, key, gEntry)
	if err != nil {
		return
	}

	checkResults = append(checkResults, groupCheck.Check(sumRefs, sumSize)...)
	if doFix {
		err = ri.fix(ctx, key, checkResults)
	}
	return
}

func (ri *redisIndex) fix(ctx context.Context, key Key, checkResults []CheckResult) (err error) {
	for _, check := range checkResults {
		switch {
		case check.SpaceId != "":
			if check.Key == infoKey {
				if err = ri.fixSpaceEntry(ctx, key, check); err != nil {
					return
				}
			} else {
				if err = ri.fixSpaceCid(ctx, key, check); err != nil {
					return
				}
			}
		case check.CidEntry != nil:
			if err = ri.fixCid(ctx, check); err != nil {
				return
			}
		case check.GroupEntry != nil:
			if check.Key == infoKey {
				if err = ri.fixGroupEntry(ctx, key, check); err != nil {
					return
				}
			} else {
				if err = ri.fixGroupCid(ctx, key, check); err != nil {
					return
				}
			}
		}
	}
	return
}

func (ri *redisIndex) fixSpaceEntry(ctx context.Context, key Key, check CheckResult) (err error) {
	se := check.SpaceEntry
	se.UpdateTime = time.Now().Unix()
	data, err := se.Marshal()
	if err != nil {
		return
	}
	return ri.cl.HSet(ctx, SpaceKey(Key{GroupId: key.GroupId, SpaceId: check.SpaceId}), infoKey, data).Err()
}

func (ri *redisIndex) fixSpaceCid(ctx context.Context, key Key, check CheckResult) (err error) {
	if check.CidRef > 0 {
		return ri.cl.HSet(ctx, SpaceKey(Key{GroupId: key.GroupId, SpaceId: check.SpaceId}), check.Key, check.CidRef).Err()
	} else {
		return ri.cl.HDel(ctx, SpaceKey(Key{GroupId: key.GroupId, SpaceId: check.SpaceId}), check.Key).Err()
	}
}

func (ri *redisIndex) fixGroupEntry(ctx context.Context, key Key, check CheckResult) (err error) {
	ge := &groupEntry{
		GroupEntry: check.GroupEntry,
	}
	ge.Save(ctx, ri.cl)
	return
}

func (ri *redisIndex) fixGroupCid(ctx context.Context, key Key, check CheckResult) (err error) {
	if check.CidRef > 0 {
		return ri.cl.HSet(ctx, GroupKey(key), check.Key, check.CidRef).Err()
	} else {
		return ri.cl.HDel(ctx, GroupKey(key), check.Key).Err()
	}
}

func (ri *redisIndex) fixCid(ctx context.Context, v CheckResult) (err error) {
	ce := v.CidEntry
	ce.UpdateTime = time.Now().Unix()
	data, err := ce.Marshal()
	if err != nil {
		return err
	}
	return ri.cl.Set(ctx, v.Key, data, 0).Err()
}

func (ri *redisIndex) loadSpaceContent(ctx context.Context, key Key, se *spaceEntry) (sc *spaceContent, err error) {
	sc = &spaceContent{
		entry:      se,
		files:      make(map[string]*indexproto.FileEntry),
		cids:       make(map[string]uint64),
		cidEntries: make(map[string]*cidEntry),
	}
	results, err := ri.cl.HGetAll(ctx, SpaceKey(key)).Result()
	if err != nil {
		return
	}
	for k, v := range results {
		if strings.HasPrefix(k, "c:") {
			sc.cids[k[2:]], _ = strconv.ParseUint(v, 10, 64)
		} else if strings.HasPrefix(k, "f:") {
			fileEntryProto := &indexproto.FileEntry{}
			if err = fileEntryProto.Unmarshal([]byte(v)); err != nil {
				return
			}
			sc.files[k[2:]] = fileEntryProto
		}
	}
	return
}

type spaceContent struct {
	entry      *spaceEntry
	files      map[string]*indexproto.FileEntry
	cids       map[string]uint64
	cidEntries map[string]*cidEntry
	actualRefs map[string]uint64
}

type CheckResult struct {
	Key         string                 `json:"key"`
	CidEntry    *indexproto.CidEntry   `json:"cid,omitempty"`
	FileEntry   *indexproto.FileEntry  `json:"file,omitempty"`
	SpaceEntry  *indexproto.SpaceEntry `json:"space,omitempty"`
	GroupEntry  *indexproto.GroupEntry `json:"group,omitempty"`
	CidRef      uint64                 `json:"cidRef,omitempty"`
	Description string                 `json:"description"`
	SpaceId     string                 `json:"spaceId,omitempty"`
}

func (sc *spaceContent) Check(ctx context.Context, ri *redisIndex) (checkResults []CheckResult, err error) {
	sc.actualRefs = make(map[string]uint64)
	// calc file refs
	for _, file := range sc.files {
		for _, fCid := range file.Cids {
			sc.actualRefs[fCid] += 1
		}
	}

	// load cid entries
	cidStrings := make([]string, 0, len(sc.actualRefs))
	for cidString := range sc.actualRefs {
		cidStrings = append(cidStrings, cidString)
	}
	entries, err := ri.CidEntriesByString(ctx, cidStrings)
	if err != nil {
		return
	}
	defer entries.Release()
	for _, cEntry := range entries.entries {
		sc.cidEntries[cEntry.Cid.String()] = cEntry
	}

	// calc files sizes
	for fileId, file := range sc.files {
		var fileSize uint64
		for _, fCid := range file.Cids {
			fileSize += sc.cidEntries[fCid].Size_
		}
		if file.Size_ != fileSize {
			fix := CheckResult{
				Key:         "f:" + fileId,
				Description: fmt.Sprintf("file size mismatch: %d -> %d", file.Size_, fileSize),
				SpaceId:     sc.entry.Id,
			}
			file.Size_ = fileSize
			fix.FileEntry = file
			checkResults = append(checkResults, fix)
		}
	}

	// check cid refs
	var sumSize uint64
	for c, want := range sc.actualRefs {
		if actual := sc.cids[c]; actual != want {
			fix := CheckResult{
				Key:         "c:" + c,
				CidRef:      want,
				Description: fmt.Sprintf("space cid refs mismatch: stored: %d ->  %d", actual, want),
				SpaceId:     sc.entry.Id,
			}
			checkResults = append(checkResults, fix)
		}
		cEntry := sc.cidEntries[c]
		if cEntry.Refs < 1 {
			cEntry.CidEntry.Refs = 1
			fix := CheckResult{
				Key:         "c:" + c,
				CidEntry:    cEntry.CidEntry,
				Description: "cid 0-ref",
			}
			checkResults = append(checkResults, fix)
		}
		sumSize += cEntry.Size_
	}

	// check space entry
	if sc.entry.Size_ != sumSize || sc.entry.FileCount != uint32(len(sc.files)) || sc.entry.CidCount != uint64(len(sc.actualRefs)) {
		fix := CheckResult{
			Key: "info",
			Description: fmt.Sprintf("space entry; size: %d -> %d; cidsCount: %d -> %d; filesCount: %d -> %d",
				sc.entry.Size_, sumSize,
				sc.entry.CidCount, len(sc.actualRefs),
				sc.entry.FileCount, len(sc.files),
			),
			SpaceId: sc.entry.Id,
		}
		sc.entry.Size_ = sumSize
		sc.entry.FileCount = uint32(len(sc.files))
		sc.entry.CidCount = uint64(len(sc.actualRefs))
		fix.SpaceEntry = sc.entry.SpaceEntry
		checkResults = append(checkResults, fix)
	}

	// check for extra cids
	for c := range sc.cids {
		if _, ok := sc.actualRefs[c]; !ok {
			fix := CheckResult{
				Key:         "c:" + c,
				CidRef:      0,
				Description: "extra cid",
				SpaceId:     sc.entry.Id,
			}
			checkResults = append(checkResults, fix)
		}
	}
	return
}

type groupContent struct {
	entry *groupEntry
	cids  map[string]uint64
}

func (ri *redisIndex) loadGroupContent(ctx context.Context, key Key, ge *groupEntry) (gc *groupContent, err error) {
	gc = &groupContent{
		entry: ge,
		cids:  make(map[string]uint64),
	}
	results, err := ri.cl.HGetAll(ctx, GroupKey(key)).Result()
	if err != nil {
		return
	}
	for k, v := range results {
		if strings.HasPrefix(k, "c:") {
			gc.cids[k[2:]], _ = strconv.ParseUint(v, 10, 64)
		}
	}
	return
}

func (gc *groupContent) Check(cidRefs map[string]uint64, sumSize uint64) (checkResults []CheckResult) {
	for k, ref := range cidRefs {
		if gRef := gc.cids[k]; gRef != ref {
			fix := CheckResult{
				Key:         "c:" + k,
				GroupEntry:  gc.entry.GroupEntry,
				CidRef:      ref,
				Description: fmt.Sprintf("group ref mismatch: %d -> %d", gRef, ref),
			}
			checkResults = append(checkResults, fix)
		}
	}
	for k, ref := range gc.cids {
		if _, ok := cidRefs[k]; !ok {
			fix := CheckResult{
				Key:         "c:" + k,
				GroupEntry:  gc.entry.GroupEntry,
				CidRef:      0,
				Description: fmt.Sprintf("group ref extra: %d -> %d", ref, 0),
			}
			checkResults = append(checkResults, fix)
		}
	}
	if gc.entry.Size_ != sumSize {
		fix := CheckResult{
			Key:         "info",
			GroupEntry:  gc.entry.GroupEntry,
			Description: fmt.Sprintf("group size mismatch: %d -> %d", gc.entry.Size_, sumSize),
		}
		fix.GroupEntry.Size_ = sumSize
		fix.GroupEntry.CidCount = uint64(len(cidRefs))
		checkResults = append(checkResults, fix)
	}
	return
}
