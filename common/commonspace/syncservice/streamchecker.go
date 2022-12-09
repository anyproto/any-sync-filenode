package syncservice

import (
	"context"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/commonspace/spacesyncproto"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/net/rpc/rpcerr"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/nodeconf"
	"go.uber.org/zap"
)

type StreamChecker interface {
	CheckResponsiblePeers()
}

type streamChecker struct {
	spaceId       string
	connector     nodeconf.ConfConnector
	streamPool    StreamPool
	clientFactory spacesyncproto.ClientFactory
	log           *zap.Logger
	syncCtx       context.Context
}

func NewStreamChecker(
	spaceId string,
	connector nodeconf.ConfConnector,
	streamPool StreamPool,
	clientFactory spacesyncproto.ClientFactory,
	syncCtx context.Context,
	log *zap.Logger) StreamChecker {
	return &streamChecker{
		spaceId:       spaceId,
		connector:     connector,
		streamPool:    streamPool,
		clientFactory: clientFactory,
		log:           log,
		syncCtx:       syncCtx,
	}
}

func (s *streamChecker) CheckResponsiblePeers() {
	var (
		activeNodeIds []string
		configuration = s.connector.Configuration()
	)
	for _, nodeId := range configuration.NodeIds(s.spaceId) {
		if s.streamPool.HasActiveStream(nodeId) {
			s.log.Debug("has active stream for", zap.String("id", nodeId))
			activeNodeIds = append(activeNodeIds, nodeId)
			continue
		}
	}
	newPeers, err := s.connector.DialInactiveResponsiblePeers(s.syncCtx, s.spaceId, activeNodeIds)
	if err != nil {
		s.log.Error("failed to dial peers", zap.Error(err))
		return
	}

	for _, p := range newPeers {
		stream, err := s.clientFactory.Client(p).Stream(s.syncCtx)
		if err != nil {
			err = rpcerr.Unwrap(err)
			s.log.Error("failed to open stream", zap.Error(err))
			// so here probably the request is failed because there is no such space,
			// but diffService should handle such cases by sending pushSpace
			continue
		}
		// sending empty message for the server to understand from which space is it coming
		err = stream.Send(&spacesyncproto.ObjectSyncMessage{SpaceId: s.spaceId})
		if err != nil {
			err = rpcerr.Unwrap(err)
			s.log.Error("failed to send first message to stream", zap.Error(err))
			continue
		}
		err = s.streamPool.AddAndReadStreamAsync(stream)
		if err != nil {
			s.log.Error("failed to read from stream async", zap.Error(err))
			continue
		}
		s.log.Debug("reading stream for", zap.String("id", p.Id()))
	}
	return
}
