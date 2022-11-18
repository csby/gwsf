package gcluster

import (
	"crypto/tls"
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"time"
)

func NewController(log gtype.Log, cfg *gcfg.Config, chs *Channels) *Controller {
	inst := &Controller{}
	inst.SetLog(log)
	inst.cfg = cfg
	inst.chs = chs
	inst.wsGrader = websocket.Upgrader{CheckOrigin: inst.checkOrigin}
	inst.inChs = gtype.NewSocketChannelCollection()

	if chs != nil {
		cluster := chs.Cluster()
		if cluster != nil {
			cluster.AddWriter(inst)
		}
	}

	if cfg != nil {
		inst.index = cfg.Cluster.Index
		inst.init(&cfg.Cluster)
	}

	return inst
}

type Controller struct {
	gtype.Base

	cfg *gcfg.Config
	chs *Channels

	wsGrader  websocket.Upgrader
	instances []*Instance
	inChs     gtype.SocketChannelCollection
	index     uint64
}

func (s *Controller) preHandle(ctx gtype.Context, ps gtype.Params) {
}

func (s *Controller) checkOrigin(r *http.Request) bool {
	if r != nil {
	}

	return true
}

func (s *Controller) createCatalog(doc gtype.Doc, names ...string) gtype.Catalog {
	root := doc.AddCatalog(ApiCatalog)

	count := len(names)
	if count < 1 {
		return root
	}

	child := root
	for i := 0; i < count; i++ {
		name := names[i]
		child = child.AddChild(name)
	}

	return child
}

func (s *Controller) createOptCatalog(doc gtype.Doc, names ...string) gtype.Catalog {
	root := doc.AddCatalog("管理平台接口")

	count := len(names)
	if count < 1 {
		return root
	}

	child := root
	for i := 0; i < count; i++ {
		name := names[i]
		child = child.AddChild(name)
	}

	return child
}

func (s *Controller) writeOptSocketMessage(id int, data interface{}) bool {
	if s.chs == nil {
		return false
	}
	if s.chs.opt == nil {
		return false
	}

	msg := &gtype.SocketMessage{
		ID:   id,
		Data: data,
	}

	s.chs.opt.Write(msg, nil)

	return true
}

func (s *Controller) init(cluster *gcfg.Cluster) {
	s.instances = make([]*Instance, 0)
	if cluster == nil {
		return
	}

	version := ""
	if s.cfg != nil {
		version = s.cfg.Module.Version
	}

	items := cluster.Instances
	c := len(items)
	for i := 0; i < c; i++ {
		item := items[i]

		if item.Index == s.index {
			continue
		}

		scheme := "ws"
		cfg := &tls.Config{
			InsecureSkipVerify: true,
		}
		if item.Secure {
			scheme = "wss"
			if cluster.Enable {
				crt := &Certificate{}
				crt.SetLog(s.GetLog())
				crt.Load(&item)
				cfg = crt.ClientTlsConfig()
			}
		}

		dialer := &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
			TLSClientConfig:  cfg,
		}
		u := url.URL{
			Scheme:   scheme,
			Host:     fmt.Sprintf("%s:%d", item.Address, item.Port),
			Path:     fmt.Sprintf("%s/instance/connect", ApiPath),
			RawQuery: fmt.Sprintf("index=%d&version=%s", item.Index, version),
		}

		channel := s.inChs.NewChannel(&gtype.Token{ID: fmt.Sprint(item.Index)})
		instance := &Instance{
			Index: item.Index,
			In: &Connection{
				Index:   item.Index,
				Channel: channel,
				Readers: s.chs.cluster,
			},
			Out: &Connection{
				Index:   item.Index,
				Channel: channel,
				Url:     u.String(),
				Host:    u.Host,
				Dialer:  dialer,
			},
		}
		instance.In.State = instance.onConnStatusChanged
		instance.Out.State = instance.onConnStatusChanged
		instance.ConnState = s.onConnStatusChanged

		s.instances = append(s.instances, instance)
	}
}

func (s *Controller) getInstance(index uint64) *Instance {
	items := s.instances
	c := len(items)
	for i := 0; i < c; i++ {
		item := items[i]
		if item == nil {
			continue
		}
		if item.Index == index {
			return item
		}
	}

	return nil
}

func (s *Controller) onConnStatusChanged(inst *Instance) {
	if inst == nil {
		return
	}

	msg := &gtype.ClusterNodeStatus{
		Index: inst.Index,
	}
	if inst.In != nil {
		msg.In = inst.In.Connected()
	}
	if inst.Out != nil {
		msg.Out = inst.Out.Connected()
	}

	go s.writeOptSocketMessage(gtype.WSClusterNodeStatusChanged, msg)
}
