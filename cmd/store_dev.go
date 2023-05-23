//go:build dev

package main

import (
	"github.com/anyproto/any-sync-filenode/store/filedevstore"
	"github.com/anyproto/any-sync/app"
)

func store() app.Component {
	return filedevstore.New()
}
