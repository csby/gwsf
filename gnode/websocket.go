package gnode

import "github.com/csby/gwsf/gtype"

type Channels struct {
	opt  gtype.SocketChannelCollection
	node gtype.SocketChannelCollection
}

func (s *Channels) Opt() gtype.SocketChannelCollection {
	return s.opt
}

func (s *Channels) Node() gtype.SocketChannelCollection {
	return s.node
}
