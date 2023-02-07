package redisindex

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEntry_Binary(t *testing.T) {
	e := Entry{
		Time: time.Now(),
		Size: 256 * 1024,
	}
	data := e.Binary()
	ne, err := NewEntry(data)
	require.NoError(t, err)
	assert.Equal(t, e.Time.Unix(), ne.Time.Unix())
	assert.Equal(t, e.Size, ne.Size)
}
