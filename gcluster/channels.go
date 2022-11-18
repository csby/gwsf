package gcluster

import "github.com/csby/gwsf/gtype"

type Channels struct {
	opt     gtype.SocketChannelCollection
	cluster gtype.SocketChannelCollection
}

func (s *Channels) Opt() gtype.SocketChannelCollection {
	return s.opt
}

func (s *Channels) Cluster() gtype.SocketChannelCollection {
	return s.cluster
}
