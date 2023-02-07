package redisindex

import (
	"bytes"
	"encoding/binary"
	"time"
)

type Entry struct {
	Time time.Time
	Size uint64
}

func (e Entry) Binary() []byte {
	var b = make([]byte, 0, 8)
	b = binary.AppendUvarint(b, uint64(e.Time.Unix()))
	b = binary.AppendUvarint(b, e.Size)
	return b
}

func NewEntry(data []byte) (e Entry, err error) {
	rd := bytes.NewReader(data)
	tm, err := binary.ReadUvarint(rd)
	if err != nil {
		return
	}
	sz, err := binary.ReadUvarint(rd)
	if err != nil {
		return
	}
	return Entry{
		Time: time.Unix(int64(tm), 0),
		Size: sz,
	}, nil
}
