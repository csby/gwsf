package gnode

import (
	"github.com/csby/gwsf/gfwd"
	"github.com/csby/gwsf/gtype"
)

func (s *innerForward) forwardTcp(request *gtype.ForwardRequest) {
	if request == nil {
		return
	}

	if request.NodeInstanceID != s.cfg.Node.InstanceId {
		return
	}

	output := &gfwd.Output{
		Remote: s.cfg.Node.CloudServer,
	}
	request.CopyTo(&output.Target)
	output.SetLog(s.GetLog())
	output.Dialer = s.dialer

	output.Start()
}
