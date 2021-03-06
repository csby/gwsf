package gfwd

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
	"net"
)

type InputUdp struct {
	Input

	Remote func(id int, data interface{}) bool

	c net.PacketConn
}

func (s *InputUdp) Start() {
	go s.run()
}

func (s *InputUdp) Stop() {
	s.close()
}

func (s *InputUdp) WriteTo(p []byte, addr string) (int, error) {
	if s.c == nil {
		return 0, fmt.Errorf("conn not opened")
	}

	a, e := net.ResolveUDPAddr("udp", addr)
	if e != nil {
		return 0, e
	}

	return s.c.WriteTo(p, a)
}

func (s *InputUdp) forwardConnect(addr string, data []byte) {
	if len(addr) < 1 {
		return
	}
	if len(data) < 1 {
		return
	}
	if s.Remote == nil {
		return
	}

	msg := &gtype.ForwardUdpRequest{}
	msg.ID = s.Local.ID
	msg.SourceNodeInstanceID = s.InstanceID
	msg.SourceAddress = addr
	msg.TargetNodeCertificateID = s.Local.TargetNodeID
	msg.TargetAddress = fmt.Sprintf("%s:%d", s.Local.TargetAddress, s.Local.TargetPort)
	msg.Data = data

	s.Remote(gtype.WSNodeForwardUdpRequest, msg)
}

func (s *InputUdp) run() {
	defer func() {
		if err := recover(); err != nil {
			s.LogError("fwd udp input listen error: ", err)
		}
	}()

	addr := fmt.Sprintf("%s:%d", s.Local.ListenAddress, s.Local.ListenPort)
	c, err := net.ListenPacket("udp", addr)
	if err != nil {
		s.LogError("fwd udp input listen fail: ", err)
		return
	}
	s.c = c
	defer s.close()

	s.LogInfo(fmt.Sprintf("forward udp input(id=%s) is ready, router: '%s' => '%s(%s)' => '%s:%d'",
		s.Local.ID,
		addr,
		s.Local.TargetNodeName, s.Local.TargetNodeID,
		s.Local.TargetAddress, s.Local.TargetPort))

	size := 65535 - 20 - 8
	for {

		b := make([]byte, size)
		n, a, e := c.ReadFrom(b)
		if e != nil {
			return
		}
		if n < 1 {
			continue
		}

		go s.forwardConnect(a.String(), b[0:n])
	}
}

func (s *InputUdp) close() {
	if s.c == nil {
		return
	}

	s.c.Close()
	s.c = nil

	s.LogInfo(fmt.Sprintf("forward udp input(id=%s) is closed, router: '%s:%d' => '%s(%s)' => '%s:%d'",
		s.Local.ID,
		s.Local.ListenAddress, s.Local.ListenPort,
		s.Local.TargetNodeName, s.Local.TargetNodeID,
		s.Local.TargetAddress, s.Local.TargetPort))
}
