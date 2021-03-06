package gnode

import (
	"github.com/csby/gwsf/gtype"
	"io"
	"net"
)

func (s *innerForward) forwardUdp(request *gtype.ForwardUdpRequest) {
	if request == nil {
		return
	}
	if request.TargetNodeInstanceID != s.cfg.Node.InstanceId {
		return
	}
	if len(request.TargetAddress) < 1 {
		return
	}
	if len(request.Data) < 1 {
		return
	}

	conn, err := net.Dial("udp", request.TargetAddress)
	if err != nil {
		return
	}
	defer conn.Close()

	_, ew := conn.Write(request.Data)
	if ew != nil {
		return
	}

	size := 65535 - 20 - 8
	buf := make([]byte, size)
	nr, er := conn.Read(buf)
	if er != nil {
		if er != io.EOF {
		}
	}
	if nr < 1 {
		return
	}

	response := &gtype.ForwardUdpResponse{}
	response.ID = request.ID
	response.SourceNodeInstanceID = request.SourceNodeInstanceID
	response.SourceAddress = request.SourceAddress
	response.Data = buf[0:nr]

	s.writeNodeMessage(gtype.WSNodeForwardUdpResponse, response)
}

func (s *innerForward) writeUdp(response *gtype.ForwardUdpResponse) {
	if response == nil {
		return
	}
	if response.SourceNodeInstanceID != s.cfg.Node.InstanceId {
		return
	}
	if len(response.SourceAddress) < 1 {
		return
	}
	if len(response.Data) < 1 {
		return
	}

	input := s.GetInputUdp(response.ID)
	if input == nil {
		return
	}

	input.WriteTo(response.Data, response.SourceAddress)
}
