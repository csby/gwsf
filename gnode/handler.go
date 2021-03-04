package gnode

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

type Handler interface {
	Init()
}

func NewHandler(log gtype.Log, cfg *gcfg.Config, optChannels gtype.SocketChannelCollection) Handler {
	instance := &innerHandler{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.optChannels = optChannels

	instance.cloud = NewCloud(log, cfg)

	return instance
}

type innerHandler struct {
	gtype.Base

	cfg         *gcfg.Config
	optChannels gtype.SocketChannelCollection

	cloud Cloud
}

func (s *innerHandler) Init() {
	err := s.cloud.Connect()
	if err != nil {
		s.LogError("connect to cloud fail: ", err)
	}
}
