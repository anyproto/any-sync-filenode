package index

import (
	"archive/zip"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisIndex_Migrate(t *testing.T) {
	fx := newFixture(t)
	defer fx.Finish(t)

	// load old space keys
	zrd, err := zip.OpenReader("testdata/oldspace.zip")
	require.NoError(t, err)
	defer zrd.Close()
	for _, zf := range zrd.File {
		zfr, err := zf.Open()
		require.NoError(t, err)
		data, err := io.ReadAll(zfr)
		require.NoError(t, err)
		require.NoError(t, fx.cl.Restore(ctx, zf.Name, 0, string(data)).Err())
		_ = zfr.Close()
	}

	// migrate
	expectedSize := uint64(18248267)
	key := newRandKey()
	key.SpaceId = "bafyreic65hvluhooz7u43hniptb4uokmpaqm6b2aneym77pivurjt4csze.2e2j1mpearah"
	require.NoError(t, fx.Migrate(ctx, key))

	spaceInfo, err := fx.SpaceInfo(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, expectedSize, spaceInfo.BytesUsage)
}
