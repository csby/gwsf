package gnode

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type Handler interface {
	Init(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle)
	Cloud() Cloud
	Forward() Forward
}

func NewHandler(log gtype.Log, cfg *gcfg.Config, optChannels gtype.SocketChannelCollection) Handler {
	instance := &innerHandler{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.chs = &Channels{
		opt:  optChannels,
		node: gtype.NewSocketChannelCollection(),
	}

	crt := &Certificate{}
	crt.SetLog(log)
	crt.Load(&cfg.Node.Cert)
	instance.dialer = &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig:  crt.ClientTlsConfig(),
	}

	instance.cloud = NewCloud(log, cfg, instance.dialer, instance.chs)
	instance.forward = NewForward(log, cfg, instance.dialer, instance.chs)
	instance.controller = &innerController{
		cfg:  cfg,
		wcs:  optChannels,
		node: instance,
	}
	if cfg != nil {
		instance.controller.nodeInstance = cfg.Node.InstanceId
	}
	if crt.node != nil {
		instance.controller.nodeId = crt.node.OrganizationalUnit()
		instance.controller.nodeName = crt.node.CommonName()
	}
	instance.forward.SetState(instance.controller.onNodeFwdInputListenStateChanged)
	instance.cloud.SetState(instance.controller.onNodeOnlineStateChanged)
	instance.SetLog(log)

	return instance
}

type innerHandler struct {
	gtype.Base

	cfg    *gcfg.Config
	chs    *Channels
	dialer *websocket.Dialer

	cloud      Cloud
	forward    Forward
	controller *innerController
}

func (s *innerHandler) Init(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle) {
	err := s.cloud.Connect()
	if err != nil {
		s.LogError("connect to cloud fail: ", err)
	}

	s.forward.Start()

	if router != nil && path != nil {
		s.controller.initRouter(router, path, preHandle)
	}
}

func (s *innerHandler) Cloud() Cloud {
	return s.cloud
}

func (s *innerHandler) Forward() Forward {
	return s.forward
}
