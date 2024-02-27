package index

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// Move moves space to the new group
func (ri *redisIndex) Move(ctx context.Context, oldKey, newKey Key) (err error) {
	if oldKey.GroupId == newKey.GroupId {
		return nil
	}

	// load old space
	oldEntry, oldRelease, err := ri.AcquireSpace(ctx, oldKey)
	if err != nil {
		return err
	}
	defer oldRelease()

	// space doesn't exist - do nothing
	if !oldEntry.spaceExists {
		return nil
	}

	// it's almost impossible, but we can get a hash collision here, just return an error and maybe think about this later
	if spaceKey(oldKey) == spaceKey(newKey) {
		log.Error("space move: hash collision", zap.String("oldGroup", oldKey.GroupId), zap.String("newGroup", newKey.GroupId), zap.String("spaceId", oldKey.SpaceId))
		return fmt.Errorf("unable to move space: hash collision")
	}

	// load new space
	newEntry, newRelease, err := ri.AcquireSpace(ctx, newKey)
	if err != nil {
		return
	}
	defer newRelease()

	// fetch and bind files
	keys, err := ri.cl.HKeys(ctx, spaceKey(oldKey)).Result()
	if err != nil {
		return
	}

	var data = moveFileData{
		oldKey:   oldKey,
		newKey:   newKey,
		oldEntry: oldEntry,
		newEntry: newEntry,
	}
	for _, k := range keys {
		if strings.HasPrefix(k, "f:") {
			if err = ri.moveFile(ctx, data, k[2:]); err != nil {
				return
			}
		}
	}

	// delete space
	if _, err = ri.spaceDelete(ctx, oldKey, oldEntry); err != nil {
		return
	}
	return
}

type moveFileData struct {
	oldKey, newKey     Key
	oldEntry, newEntry groupSpaceEntry
}

func (ri *redisIndex) moveFile(ctx context.Context, data moveFileData, fileId string) (err error) {
	fileInfo, isNewFile, err := ri.getFileEntry(ctx, data.oldKey, fileId)
	if err != nil {
		return
	}
	if isNewFile {
		return
	}

	cids, err := ri.CidEntriesByString(ctx, fileInfo.Cids)
	if err != nil {
		return err
	}
	defer cids.Release()

	return ri.fileBind(ctx, data.newKey, fileId, cids, data.newEntry)
}
