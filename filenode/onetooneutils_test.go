package filenode

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOneToOneSpaceId(t *testing.T) {
	type tc struct {
		name          string
		myKey         []byte
		pub0, pub1    []byte
		spaceId       string
		want          string
		wantSwapOrder [][]byte // expected pubKeys order after call; nil = don't check
		wantErr       bool
	}

	cp := func(b []byte) []byte { return append([]byte(nil), b...) }

	tests := []tc{
		{
			name:    "error_when_not_two_pubkeys_zero",
			myKey:   []byte{0x01},
			spaceId: "x",
			wantErr: true,
		},
		{
			name:    "error_when_not_two_pubkeys_three",
			myKey:   []byte{0x01},
			spaceId: "x",
			wantErr: true,
			// pubkeys provided below when running
		},
		{
			name:          "already_sorted_my_is_first_suffix_0",
			myKey:         []byte{0x01, 0x02},
			pub0:          []byte{0x01, 0x02},
			pub1:          []byte{0x03, 0x04},
			spaceId:       "space",
			want:          "space#0",
			wantSwapOrder: [][]byte{{0x01, 0x02}, {0x03, 0x04}},
		},
		{
			name:          "needs_swap_my_is_smaller_becomes_index0_suffix_0",
			myKey:         []byte{0x01, 0x02},
			pub0:          []byte{0x09, 0x09}, // > pub1, will swap
			pub1:          []byte{0x01, 0x02},
			spaceId:       "s",
			want:          "s#0",
			wantSwapOrder: [][]byte{{0x01, 0x02}, {0x09, 0x09}},
		},
		{
			name:          "needs_swap_my_is_larger_becomes_index1_suffix_1",
			myKey:         []byte{0x09, 0x09},
			pub0:          []byte{0x09, 0x09},
			pub1:          []byte{0x01, 0x02}, // smaller → becomes index 0
			spaceId:       "id",
			want:          "id#1",
			wantSwapOrder: [][]byte{{0x01, 0x02}, {0x09, 0x09}},
		},
		{
			name:          "myKey_not_in_pubKeys_suffix_1",
			myKey:         []byte{0xAA},
			pub0:          []byte{0x01},
			pub1:          []byte{0x02},
			spaceId:       "x",
			want:          "x#1",
			wantSwapOrder: [][]byte{{0x01}, {0x02}},
		},
		{
			name:          "equal_pubkeys_my_matches_suffix_0",
			myKey:         []byte{0x11},
			pub0:          []byte{0x11},
			pub1:          []byte{0x11},
			spaceId:       "eq",
			want:          "eq#0",
			wantSwapOrder: [][]byte{{0x11}, {0x11}},
		},
		{
			name:          "equal_pubkeys_my_diff_suffix_1",
			myKey:         []byte{0x22},
			pub0:          []byte{0x11},
			pub1:          []byte{0x11},
			spaceId:       "eq2",
			want:          "eq2#1",
			wantSwapOrder: [][]byte{{0x11}, {0x11}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "error_when_not_two_pubkeys_zero":
				// no pubkeys
				got, err := oneToOneSpaceId(cp(tt.myKey), nil, tt.spaceId)
				if err == nil {
					t.Fatalf("expected error, got result %q", got)
				}
				return
			case "error_when_not_two_pubkeys_three":
				pubKeys := [][]byte{{1}, {2}, {3}}
				got, err := oneToOneSpaceId([]byte{1}, pubKeys, "x")
				if err == nil {
					t.Fatalf("expected error, got result %q", got)
				}
				return
			}

			pubKeys := [][]byte{cp(tt.pub0), cp(tt.pub1)} // defensive copy

			got, err := oneToOneSpaceId(cp(tt.myKey), pubKeys, tt.spaceId)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("oneToOneSpaceId() = %q, want %q", got, tt.want)
			}
			if tt.wantSwapOrder != nil {
				if !bytes.Equal(pubKeys[0], tt.wantSwapOrder[0]) || !bytes.Equal(pubKeys[1], tt.wantSwapOrder[1]) {
					t.Fatalf("pubKeys mutated to unexpected order: got [%v %v], want [%v %v]",
						pubKeys[0], pubKeys[1], tt.wantSwapOrder[0], tt.wantSwapOrder[1])
				}
			}
		})
	}
}

// Documents the side effect clearly: pubKeys are sorted in-place.
func TestOneToOneSpaceId_MutatesPubKeysOrder(t *testing.T) {
	p0 := []byte{0xFF}
	p1 := []byte{0x00}
	pubKeys := [][]byte{append([]byte(nil), p0...), append([]byte(nil), p1...)} // unsorted

	_, err := oneToOneSpaceId(p0, pubKeys, "z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(pubKeys[0], p1) || !bytes.Equal(pubKeys[1], p0) {
		t.Fatalf("expected in-place swap; got [%v %v], want [%v %v]", pubKeys[0], pubKeys[1], p1, p0)
	}
}

// Optional: a tiny property—result must be order-invariant except for suffix flip.
func TestOneToOneSpaceId_OrderInvariance(t *testing.T) {
	my := []byte{0xAA}
	a := []byte{0x01}
	b := []byte{0x02}

	got1, err1 := oneToOneSpaceId(my, [][]byte{append([]byte(nil), a...), append([]byte(nil), b...)}, "s")
	if err1 != nil {
		t.Fatalf("err1: %v", err1)
	}
	got2, err2 := oneToOneSpaceId(my, [][]byte{append([]byte(nil), b...), append([]byte(nil), a...)}, "s")
	if err2 != nil {
		t.Fatalf("err2: %v", err2)
	}

	// Both must end with "#1" because my != min(a,b), regardless of input order.
	if got1 != "s#1" || got2 != "s#1" {
		t.Fatalf("order invariance violated: got %q and %q, want 's#1' both", got1, got2)
	}
}

func Test_oneToOneParticipantPubKeys(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)
	spaceId := "spaceId"
	aclList := oneToOneAclList(t, spaceId)
	accounts := aclList.AclState().CurrentAccounts()

	t.Run("basic test", func(t *testing.T) {
		writersPubKeys, err := oneToOneParticipantPubKeys(accounts)

		require.NoError(t, err)
		assert.False(t, bytes.Equal(writersPubKeys[0], writersPubKeys[1]))

		ownerPubKey, _ := aclList.AclState().OwnerPubKey()
		ownerBytes := ownerPubKey.Storage()
		assert.False(t, bytes.Equal(writersPubKeys[0], ownerBytes))
		assert.False(t, bytes.Equal(writersPubKeys[1], ownerBytes))

	})

	t.Run("basic test", func(t *testing.T) {
		accounts = append(accounts, accounts[1])
		_, err := oneToOneParticipantPubKeys(accounts)
		require.Error(t, err)

	})
}
