//go:build dev

package main

import (
	"github.com/anytypeio/any-sync-filenode/filedevstore"
	"github.com/anytypeio/any-sync/app"
)

func store() app.Component {
	return filedevstore.New()
}
