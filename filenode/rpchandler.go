package filenode

import (
	"context"
	"time"

	"github.com/anyproto/any-sync/commonfile/fileproto"
	"github.com/anyproto/any-sync/commonfile/fileproto/fileprotoerr"
	"github.com/anyproto/any-sync/metric"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"go.uber.org/zap"
)

const (
	cidSizeLimit     = 2 << 20 // 2 Mb
	fileInfoReqLimit = 1000
)

type rpcHandler struct {
	f *fileNode
}

func (r rpcHandler) BlockGet(ctx context.Context, req *fileproto.BlockGetRequest) (resp *fileproto.BlockGetResponse, err error) {
	var c cid.Cid
	st := time.Now()
	defer func() {
		var size int
		if resp != nil {
			size = len(resp.Data)
		}
		r.f.metric.RequestLog(ctx,
			"file.blockGet",
			metric.TotalDur(time.Since(st)),
			metric.SpaceId(req.SpaceId),
			metric.Size(size),
			metric.Cid(c.String()),
			zap.Bool("wait", req.Wait),
			zap.Error(err),
		)
	}()
	resp = &fileproto.BlockGetResponse{
		Cid: req.Cid,
	}
	c, err = cid.Cast(req.Cid)
	if err != nil {
		return nil, err
	}
	b, err := r.f.Get(ctx, c, req.Wait)
	if err != nil {
		return nil, err
	} else {
		resp.Data = b.RawData()
	}
	return resp, nil
}

func (r rpcHandler) BlockPush(ctx context.Context, req *fileproto.BlockPushRequest) (resp *fileproto.Ok, err error) {
	var c cid.Cid
	st := time.Now()
	defer func() {
		r.f.metric.RequestLog(ctx,
			"file.blockPush",
			metric.TotalDur(time.Since(st)),
			metric.SpaceId(req.SpaceId),
			metric.Size(len(req.Data)),
			metric.Cid(c.String()),
			metric.FileId(req.FileId),
			zap.Error(err),
		)
	}()
	c, err = cid.Cast(req.Cid)
	if err != nil {
		return nil, err
	}

	b, err := blocks.NewBlockWithCid(req.Data, c)
	if err != nil {
		return nil, err
	}
	if len(req.Data) > cidSizeLimit {
		return nil, fileprotoerr.ErrQuerySizeExceeded
	}
	chkc, err := c.Prefix().Sum(req.Data)
	if err != nil {
		return nil, err
	}
	if !chkc.Equals(c) {
		return nil, ErrWrongHash
	}

	if err = r.f.Add(ctx, req.SpaceId, req.FileId, []blocks.Block{b}); err != nil {
		return nil, err
	}
	return &fileproto.Ok{}, nil
}

func (r rpcHandler) BlocksCheck(ctx context.Context, req *fileproto.BlocksCheckRequest) (resp *fileproto.BlocksCheckResponse, err error) {
	st := time.Now()
	defer func() {
		r.f.metric.RequestLog(ctx,
			"file.blocksCheck",
			metric.TotalDur(time.Since(st)),
			metric.SpaceId(req.SpaceId),
			metric.Size(len(req.Cids)),
			zap.Error(err),
		)
	}()
	availability, err := r.f.Check(ctx, req.SpaceId, convertCids(req.Cids)...)
	if err != nil {
		return nil, err
	}
	return &fileproto.BlocksCheckResponse{
		BlocksAvailability: availability,
	}, nil
}

func (r rpcHandler) BlocksBind(ctx context.Context, req *fileproto.BlocksBindRequest) (resp *fileproto.Ok, err error) {
	st := time.Now()
	defer func() {
		r.f.metric.RequestLog(ctx,
			"file.blocksBind",
			metric.TotalDur(time.Since(st)),
			metric.SpaceId(req.SpaceId),
			metric.FileId(req.FileId),
			metric.Size(len(req.Cids)),
			zap.Error(err),
		)
	}()
	if err = r.f.BlocksBind(ctx, req.SpaceId, req.FileId, convertCids(req.Cids)...); err != nil {
		return nil, err
	}
	return &fileproto.Ok{}, nil
}

func (r rpcHandler) FilesDelete(ctx context.Context, req *fileproto.FilesDeleteRequest) (resp *fileproto.FilesDeleteResponse, err error) {
	st := time.Now()
	defer func() {
		r.f.metric.RequestLog(ctx,
			"file.filesDelete",
			metric.TotalDur(time.Since(st)),
			metric.SpaceId(req.SpaceId),
			metric.Size(len(req.FileIds)),
			zap.Error(err),
		)
	}()
	if err = r.f.FilesDelete(ctx, req.SpaceId, req.FileIds); err != nil {
		return nil, err
	}
	return &fileproto.FilesDeleteResponse{}, nil
}

func (r rpcHandler) FilesInfo(ctx context.Context, req *fileproto.FilesInfoRequest) (resp *fileproto.FilesInfoResponse, err error) {
	st := time.Now()
	defer func() {
		r.f.metric.RequestLog(ctx,
			"file.filesInfo",
			metric.TotalDur(time.Since(st)),
			metric.SpaceId(req.SpaceId),
			metric.Size(len(req.FileIds)),
			zap.Error(err),
		)
	}()
	if len(req.FileIds) > fileInfoReqLimit {
		err = fileprotoerr.ErrQuerySizeExceeded
		return
	}
	resp = &fileproto.FilesInfoResponse{}
	resp.FilesInfo, err = r.f.FileInfo(ctx, req.SpaceId, req.FileIds...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (r rpcHandler) FilesGet(req *fileproto.FilesGetRequest, stream fileproto.DRPCFile_FilesGetStream) (err error) {
	ctx := stream.Context()
	st := time.Now()
	defer func() {
		r.f.metric.RequestLog(ctx,
			"file.filesGet",
			metric.TotalDur(time.Since(st)),
			metric.SpaceId(req.SpaceId),
			zap.Error(err),
		)
	}()

	fileIds, err := r.f.FilesGet(ctx, req.SpaceId)
	if err != nil {
		return
	}

	for _, fileId := range fileIds {
		if err = stream.Send(&fileproto.FilesGetResponse{
			FileId: fileId,
		}); err != nil {
			return
		}
	}
	return
}

func (r rpcHandler) Check(ctx context.Context, req *fileproto.CheckRequest) (*fileproto.CheckResponse, error) {
	st := time.Now()
	defer func() {
		r.f.metric.RequestLog(ctx,
			"file.check",
			metric.TotalDur(time.Since(st)),
		)
	}()
	return &fileproto.CheckResponse{
		SpaceIds:   nil,
		AllowWrite: true,
	}, nil
}

func (r rpcHandler) SpaceInfo(ctx context.Context, req *fileproto.SpaceInfoRequest) (resp *fileproto.SpaceInfoResponse, err error) {
	st := time.Now()
	defer func() {
		r.f.metric.RequestLog(ctx,
			"file.spaceInfo",
			metric.TotalDur(time.Since(st)),
			metric.SpaceId(req.SpaceId),
			zap.Error(err),
		)
	}()
	if resp, err = r.f.SpaceInfo(ctx, req.SpaceId); err != nil {
		return
	}
	return
}

func (r rpcHandler) AccountInfo(ctx context.Context, req *fileproto.AccountInfoRequest) (resp *fileproto.AccountInfoResponse, err error) {
	st := time.Now()
	defer func() {
		r.f.metric.RequestLog(ctx,
			"file.accountInfo",
			metric.TotalDur(time.Since(st)),
			zap.Error(err),
		)
	}()
	if resp, err = r.f.AccountInfo(ctx); err != nil {
		return
	}
	return
}

func (r rpcHandler) AccountLimitSet(ctx context.Context, req *fileproto.AccountLimitSetRequest) (resp *fileproto.Ok, err error) {
	st := time.Now()
	defer func() {
		r.f.metric.RequestLog(ctx,
			"file.accountLimitSet",
			metric.TotalDur(time.Since(st)),
			zap.Error(err),
		)
	}()
	if err = r.f.AccountLimitSet(ctx, req.Identity, req.Limit); err != nil {
		return
	}
	return &fileproto.Ok{}, nil
}

func (r rpcHandler) SpaceLimitSet(ctx context.Context, req *fileproto.SpaceLimitSetRequest) (resp *fileproto.Ok, err error) {
	st := time.Now()
	defer func() {
		r.f.metric.RequestLog(ctx,
			"file.spaceLimitSet",
			metric.TotalDur(time.Since(st)),
			zap.Error(err),
		)
	}()
	if err = r.f.SpaceLimitSet(ctx, req.SpaceId, req.Limit); err != nil {
		return
	}
	return &fileproto.Ok{}, nil
}

func convertCids(bCids [][]byte) (cids []cid.Cid) {
	cids = make([]cid.Cid, 0, len(bCids))
	var uniqMap map[string]struct{}
	if len(bCids) > 1 {
		uniqMap = make(map[string]struct{})
	}
	for _, cd := range bCids {
		c, err := cid.Cast(cd)
		if err == nil {
			if uniqMap != nil {
				if _, ok := uniqMap[c.KeyString()]; ok {
					continue
				} else {
					uniqMap[c.KeyString()] = struct{}{}
				}
			}
			cids = append(cids, c)
		}
	}
	return
}
