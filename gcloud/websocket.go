package gcloud

import (
	"github.com/csby/gwsf/gtype"
)

type Channels struct {
	opt     gtype.SocketChannelCollection
	node    gtype.SocketChannelCollection
	forward gtype.SocketChannelCollection
}

func (s *Channels) Opt() gtype.SocketChannelCollection {
	return s.opt
}

func (s *Channels) Node() gtype.SocketChannelCollection {
	return s.node
}

func (s *Channels) Forward() gtype.SocketChannelCollection {
	return s.forward
}
