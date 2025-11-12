package filenode

import (
	"bytes"
	"fmt"

	"github.com/anyproto/any-sync/commonspace/object/acl/list"
)

// Extracts []PubKey.Storage for both 1-1 participants from accounts
func oneToOneParticipantPubKeys(accounts []list.AccountState) ([][]byte, error) {
	writerPubKeys := make([][]byte, 2)
	i := 0
	if len(accounts) != 3 {
		return nil, fmt.Errorf("getOneToOnePaticipantPubKeys: expected 3 accounts, got %d", len(accounts))
	}
	for _, a := range accounts {
		// skip ephimeral owner of 1-1 space
		if a.Permissions.IsOwner() {
			continue
		}
		writerPubKeys[i] = a.PubKey.Storage()
	}
	return writerPubKeys, nil
}

// Returns determenistic spaceId for 1-1 participant.
//
// In order to account files for each 1-1 participant separately,
// we put them to an ephimeral spaceId, different for each participant.
// Therefore, each participant writes to their own space so we can
// count limits per identity.
// (pubKeys will be sorted as a side effect)
func oneToOneSpaceId(myKey []byte, pubKeys [][]byte, spaceId string) (string, error) {
	if len(pubKeys) != 2 {
		return "", fmt.Errorf("oneToOneSpaceId: expected 2 pubKeys, got %d", len(pubKeys))
	}

	if bytes.Compare(pubKeys[0], pubKeys[1]) > 0 {
		tmp := pubKeys[0]
		pubKeys[0] = pubKeys[1]
		pubKeys[1] = tmp
	}
	var suffix string
	if bytes.Equal(myKey, pubKeys[0]) {
		suffix = "#0"
	} else {
		suffix = "#1"
	}

	return fmt.Sprintf("%s%s", spaceId, suffix), nil
}
