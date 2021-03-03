package gcloud

import "github.com/csby/gwsf/gtype"

type Channels struct {
	opt          gtype.SocketChannelCollection
	onlineNodes  gtype.SocketChannelCollection
	fwdRequests  gtype.SocketChannelCollection
	fwdResponses gtype.SocketChannelCollection
}

func (s *Channels) Opt() gtype.SocketChannelCollection {
	return s.opt
}

func (s *Channels) OnlineNodes() gtype.SocketChannelCollection {
	return s.onlineNodes
}

func (s *Channels) FwdRequests() gtype.SocketChannelCollection {
	return s.fwdRequests
}

func (s *Channels) FwdResponses() gtype.SocketChannelCollection {
	return s.fwdResponses
}
