package filenode

import (
	"context"
	"github.com/anytypeio/any-sync/commonfile/fileproto"
	"github.com/anytypeio/any-sync/commonfile/fileproto/fileprotoerr"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
)

const (
	cidSizeLimit     = 2 << 20 // 2 Mb
	fileInfoReqLimit = 100
)

type rpcHandler struct {
	f *fileNode
}

func (r rpcHandler) BlockGet(ctx context.Context, req *fileproto.BlockGetRequest) (*fileproto.BlockGetResponse, error) {
	resp := &fileproto.BlockGetResponse{
		Cid: req.Cid,
	}
	c, err := cid.Cast(req.Cid)
	if err != nil {
		return nil, err
	}
	b, err := r.f.Get(ctx, c)
	if err != nil {
		return nil, err
	} else {
		resp.Data = b.RawData()
	}
	return resp, nil
}

func (r rpcHandler) BlockPush(ctx context.Context, req *fileproto.BlockPushRequest) (*fileproto.BlockPushResponse, error) {
	c, err := cid.Cast(req.Cid)
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
	return &fileproto.BlockPushResponse{}, nil
}

func (r rpcHandler) BlocksCheck(ctx context.Context, req *fileproto.BlocksCheckRequest) (*fileproto.BlocksCheckResponse, error) {
	availability, err := r.f.Check(ctx, req.SpaceId, convertCids(req.Cids)...)
	if err != nil {
		return nil, err
	}
	return &fileproto.BlocksCheckResponse{
		BlocksAvailability: availability,
	}, nil
}

func (r rpcHandler) BlocksBind(ctx context.Context, req *fileproto.BlocksBindRequest) (*fileproto.BlocksBindResponse, error) {
	if err := r.f.BlocksBind(ctx, req.SpaceId, req.FileId, convertCids(req.Cids)...); err != nil {
		return nil, err
	}
	return &fileproto.BlocksBindResponse{}, nil
}

func (r rpcHandler) FilesDelete(ctx context.Context, req *fileproto.FilesDeleteRequest) (*fileproto.FilesDeleteResponse, error) {
	if err := r.f.FilesDelete(ctx, req.SpaceId, req.FileIds); err != nil {
		return nil, err
	}
	return &fileproto.FilesDeleteResponse{}, nil
}

func (r rpcHandler) FilesInfo(ctx context.Context, req *fileproto.FilesInfoRequest) (*fileproto.FilesInfoResponse, error) {
	if len(req.FileIds) > fileInfoReqLimit {
		return nil, fileprotoerr.ErrQuerySizeExceeded
	}
	resp := &fileproto.FilesInfoResponse{
		FilesInfo: make([]*fileproto.FileInfo, len(req.FileIds)),
	}
	if err := r.f.ValidateSpaceId(ctx, req.SpaceId, false); err != nil {
		return nil, err
	}
	for i, fileId := range req.FileIds {
		info, err := r.f.FileInfo(ctx, req.SpaceId, fileId)
		if err != nil {
			return nil, err
		}
		resp.FilesInfo[i] = info
	}
	return resp, nil
}

func (r rpcHandler) Check(ctx context.Context, req *fileproto.CheckRequest) (*fileproto.CheckResponse, error) {
	return &fileproto.CheckResponse{
		SpaceIds:   nil,
		AllowWrite: true,
	}, nil
}

func (r rpcHandler) SpaceInfo(ctx context.Context, req *fileproto.SpaceInfoRequest) (*fileproto.SpaceInfoResponse, error) {
	return r.f.SpaceInfo(ctx, req.SpaceId)
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
