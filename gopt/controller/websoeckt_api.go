package controller

import "github.com/csby/gwsf/gtype"

// implement of gtype.Opt

func (s *Websocket) Tdb() gtype.TokenDatabase {
	return s.dbToken
}

func (s *Websocket) Wsc() gtype.SocketChannelCollection {
	return s.wsChannels
}

func (s *Websocket) Appendix() (input, output gtype.Appendix) {
	return s.input, s.output
}
