//go:build !dev

package main

import (
	"github.com/anyproto/any-sync-filenode/store/s3store"
	"github.com/anyproto/any-sync/app"
)

func store() app.Component {
	return s3store.New()
}
