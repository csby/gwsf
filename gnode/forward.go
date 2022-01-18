package gnode

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gfwd"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"strings"
	"sync"
)

type Forward interface {
	SetState(state func(id string, isRunning bool, lastError string, newCount, oldCount int))
	GetStates(states map[string]*gcfg.FwdState)
	IsRunning() bool
	Start()
	Stop()
}

func NewForward(log gtype.Log, cfg *gcfg.Config, dialer *websocket.Dialer, chs *Channels) Forward {
	instance := &innerForward{
		runningCount: 0,
	}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.dialer = dialer
	instance.chs = chs
	instance.tcpInputs = make([]*gfwd.InputTcp, 0)
	instance.udpInputs = make([]*gfwd.InputUdp, 0)

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

	stateMutex   sync.RWMutex
	runningCount int
	state        func(id string, isRunning bool, lastError string, newCount, oldCount int)
}

func (s *innerForward) SetState(state func(id string, isRunning bool, lastError string, newCount, oldCount int)) {
	s.state = state
}

func (s *innerForward) IsRunning() bool {
	if s.runningCount > 0 {
		return true
	} else {
		return false
	}
}

func (s *innerForward) Start() {
	s.Stop()
	s.buildForwards()

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
		input.StateChanged = s.onInputStateChanged

		s.tcpInputs = append(s.tcpInputs, input)
		input.Start()
	}

	// udp
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
		input.StateChanged = s.onInputStateChanged

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

func (s *innerForward) GetStates(states map[string]*gcfg.FwdState) {
	if states == nil {
		return
	}

	// TCP
	c := len(s.tcpInputs)
	for i := 0; i < c; i++ {
		item := s.tcpInputs[i]
		state, ok := states[item.Local.ID]
		if !ok {
			continue
		}
		state.IsRunning = item.IsRunning()
		if !state.IsRunning {
			state.LastError = item.LastError()
		}
	}

	// UDP
	c = len(s.udpInputs)
	for i := 0; i < c; i++ {
		item := s.udpInputs[i]
		state, ok := states[item.Local.ID]
		if !ok {
			continue
		}
		state.IsRunning = item.IsRunning()
		if !state.IsRunning {
			state.LastError = item.LastError()
		}
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

	items := s.cfg.Node.Forward.Items
	count := len(items)
	for index := 0; index < count; index++ {
		item := items[index]
		if item == nil {
			continue
		}
		if !item.Enable {
			continue
		}

		fwd := gcfg.Fwd{}
		item.CopyTo(&fwd)

		protocol := strings.ToLower(item.Protocol)
		if protocol == "tcp" {
			s.tcpForwards = append(s.tcpForwards, fwd)
		} else if protocol == "udp" {
			s.udpForwards = append(s.udpForwards, fwd)
		}
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

func (s *innerForward) onInputStateChanged(id string, isRunning bool, lastError string) {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()

	oldCount := s.runningCount
	if isRunning {
		s.runningCount++
	} else {
		s.runningCount--
	}
	if s.state != nil {
		s.state(id, isRunning, lastError, s.runningCount, oldCount)
	}
}
