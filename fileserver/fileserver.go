package fileserver

import (
	"github.com/anytypeio/any-sync-filenode/serverstore"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/app/logger"
	"github.com/anytypeio/any-sync/commonfile/fileblockstore"
	"github.com/anytypeio/any-sync/commonfile/fileproto"
	"github.com/anytypeio/any-sync/net/rpc/server"
)

const CName = "filenode.fileserver"

var log = logger.NewNamed(CName)

func New() FileServer {
	return &fileServer{}
}

type FileServer interface {
	app.Component
}

type fileServer struct {
	store serverstore.ServerStore
}

func (f *fileServer) Init(a *app.App) (err error) {
	f.store = a.MustComponent(fileblockstore.CName).(serverstore.ServerStore)
	return fileproto.DRPCRegisterFile(a.MustComponent(server.CName).(server.DRPCServer), &rpcHandler{store: f.store})
}

func (f *fileServer) Name() (name string) {
	return CName
}
