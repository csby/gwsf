package gnode

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gfwd"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
)

type Forward interface {
	Start()
	Stop()
}

func NewForward(log gtype.Log, cfg *gcfg.Config, dialer *websocket.Dialer, chs *Channels) Forward {
	instance := &innerForward{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.dialer = dialer
	instance.chs = chs
	instance.inputs = make([]*gfwd.Input, 0)

	instance.buildForwards()

	if chs != nil {
		if chs.node != nil {
			chs.node.AddReader(instance.readNodeMessage)
		}
	}

	return instance
}

type innerForward struct {
	gtype.Base

	cfg    *gcfg.Config
	dialer *websocket.Dialer
	chs    *Channels

	forwards []gcfg.Fwd
	inputs   []*gfwd.Input
}

func (s *innerForward) Start() {
	s.Stop()

	count := len(s.forwards)
	if count < 1 {
		return
	}

	for index := 0; index < count; index++ {
		fwd := s.forwards[index]
		if !fwd.Enable {
			continue
		}

		input := &gfwd.Input{
			Dialer: s.dialer,
			Remote: s.cfg.Node.CloudServer,
		}
		fwd.CopyTo(&input.Local)
		input.SetLog(s.GetLog())

		s.inputs = append(s.inputs, input)
		input.Start()
	}
}

func (s *innerForward) Stop() {
	count := len(s.inputs)
	if count < 1 {
		return
	}

	for index := 0; index < count; index++ {
		input := s.inputs[index]
		if input == nil {
			continue
		}

		input.Stop()
	}
}

func (s *innerForward) buildForwards() {
	s.forwards = make([]gcfg.Fwd, 0)

	if s.cfg == nil {
		return
	}

	if !s.cfg.Node.Enabled {
		return
	}
	if !s.cfg.Node.Forward.Enable {
		return
	}

	count := len(s.cfg.Node.Forward.Tcp)
	for index := 0; index < count; index++ {
		tcp := s.cfg.Node.Forward.Tcp[index]
		if tcp == nil {
			continue
		}
		if !tcp.Enable {
			continue
		}

		fwd := gcfg.Fwd{}
		tcp.CopyTo(&fwd)

		s.forwards = append(s.forwards, fwd)
	}
}

func (s *innerForward) readNodeMessage(message *gtype.SocketMessage, channel gtype.SocketChannel) {
	if message == nil {
		return
	}

	if message.ID != gtype.WSNodeForwardRequest {
		return
	}
	if message.Data == nil {
		return
	}

	fwd, ok := message.Data.(*gtype.Forward)
	if !ok {
		return
	}
	if fwd.NodeInstanceID != s.cfg.Node.InstanceId {
		return
	}

	output := &gfwd.Output{
		Remote: s.cfg.Node.CloudServer,
		Dialer: s.dialer,
	}
	fwd.CopyTo(&output.Target)
	output.SetLog(s.GetLog())

	output.Start()
}
