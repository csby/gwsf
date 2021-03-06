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
	instance.tcpInputs = make([]*gfwd.InputTcp, 0)
	instance.udpInputs = make([]*gfwd.InputUdp, 0)

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

	tcpForwards []gcfg.Fwd
	tcpInputs   []*gfwd.InputTcp

	udpForwards []gcfg.Fwd
	udpInputs   []*gfwd.InputUdp
}

func (s *innerForward) Start() {
	s.Stop()

	// tcp
	count := len(s.tcpForwards)
	for index := 0; index < count; index++ {
		fwd := s.tcpForwards[index]
		if !fwd.Enable {
			continue
		}

		input := &gfwd.InputTcp{
			Remote: s.cfg.Node.CloudServer,
		}
		fwd.CopyTo(&input.Local)
		input.SetLog(s.GetLog())
		input.Dialer = s.dialer
		input.InstanceID = s.cfg.Node.InstanceId

		s.tcpInputs = append(s.tcpInputs, input)
		input.Start()
	}

	count = len(s.udpForwards)
	for index := 0; index < count; index++ {
		fwd := s.udpForwards[index]
		if !fwd.Enable {
			continue
		}

		input := &gfwd.InputUdp{
			Remote: s.writeNodeMessage,
		}
		fwd.CopyTo(&input.Local)
		input.SetLog(s.GetLog())
		input.Dialer = s.dialer
		input.InstanceID = s.cfg.Node.InstanceId

		s.udpInputs = append(s.udpInputs, input)
		input.Start()
	}
}

func (s *innerForward) Stop() {
	count := len(s.tcpInputs)
	for index := 0; index < count; index++ {
		input := s.tcpInputs[index]
		if input == nil {
			continue
		}

		input.Stop()
	}

	count = len(s.udpInputs)
	for index := 0; index < count; index++ {
		input := s.udpInputs[index]
		if input == nil {
			continue
		}

		input.Stop()
	}
}

func (s *innerForward) buildForwards() {
	s.tcpForwards = make([]gcfg.Fwd, 0)
	s.udpForwards = make([]gcfg.Fwd, 0)

	if s.cfg == nil {
		return
	}

	if !s.cfg.Node.Enabled {
		return
	}
	if !s.cfg.Node.Forward.Enable {
		return
	}

	// tcp
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

		s.tcpForwards = append(s.tcpForwards, fwd)
	}

	// udp
	count = len(s.cfg.Node.Forward.Udp)
	for index := 0; index < count; index++ {
		tcp := s.cfg.Node.Forward.Udp[index]
		if tcp == nil {
			continue
		}
		if !tcp.Enable {
			continue
		}

		fwd := gcfg.Fwd{}
		tcp.CopyTo(&fwd)

		s.udpForwards = append(s.udpForwards, fwd)
	}
}

func (s *innerForward) GetInputUdp(id string) *gfwd.InputUdp {
	if s.udpInputs == nil {
		return nil
	}

	count := len(s.udpInputs)
	for index := 0; index < count; index++ {
		item := s.udpInputs[index]
		if item == nil {
			continue
		}
		if item.Local.ID == id {
			return item
		}
	}

	return nil
}

func (s *innerForward) writeNodeMessage(id int, data interface{}) bool {
	if s.chs == nil {
		return false
	}
	if s.chs.node == nil {
		return false
	}

	msg := &gtype.SocketMessage{
		ID:   id,
		Data: data,
	}

	s.chs.node.WriteMessage(msg, s.cfg.Node.InstanceId)

	return true
}

func (s *innerForward) readNodeMessage(message *gtype.SocketMessage, channel gtype.SocketChannel) {
	if message == nil {
		return
	}

	if message.Data == nil {
		return
	}

	if message.ID == gtype.WSNodeForwardTcpStart {
		fwd := &gtype.ForwardRequest{}
		err := message.GetData(fwd)
		if err == nil {
			go s.forwardTcp(fwd)
		}
	} else if message.ID == gtype.WSNodeForwardUdpRequest {
		fwd := &gtype.ForwardUdpRequest{}
		err := message.GetData(fwd)
		if err == nil {
			go s.forwardUdp(fwd)
		}
	} else if message.ID == gtype.WSNodeForwardUdpResponse {
		fwd := &gtype.ForwardUdpResponse{}
		err := message.GetData(fwd)
		if err == nil {
			go s.writeUdp(fwd)
		}
	}
}
