//go:build !dev

package main

import (
	"github.com/anytypeio/any-sync-filenode/s3store"
	"github.com/anytypeio/any-sync/app"
)

func store() app.Component {
	return s3store.New()
}