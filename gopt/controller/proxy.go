package controller

import (
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gmodel"
	"github.com/csby/gwsf/gproxy"
	"github.com/csby/gwsf/gtype"
	"net"
	"strconv"
	"time"
)

func NewProxy(log gtype.Log, cfg *gcfg.Config, wsc gtype.SocketChannelCollection) *Proxy {
	inst := &Proxy{}
	inst.SetLog(log)
	inst.cfg = cfg
	inst.wsChannels = wsc

	inst.proxyLinks = gproxy.NewLinkCollection()
	inst.proxyServer = &gproxy.Server{
		StatusChanged:            inst.onProxyServerStatusChanged,
		OnConnected:              inst.onProxyConnected,
		OnDisconnected:           inst.onProxyDisconnected,
		OnTargetAliveChanged:     inst.onTargetAliveChanged,
		OnTargetConnCountChanged: inst.onTargetConnCountChanged,
	}
	inst.proxyServer.SetLog(log)
	inst.proxyTargets = &ProxyTargetCollection{
		items: make(map[string]ProxyTargetItem),
	}

	inst.initRoutes()
	if len(inst.proxyServer.Routes) > 0 && cfg.ReverseProxy.Disable == false {
		inst.proxyServer.Start()
	}

	return inst

	return inst
}

type Proxy struct {
	controller

	proxyServer  *gproxy.Server
	proxyLinks   gproxy.LinkCollection
	proxyTargets *ProxyTargetCollection
}

func (s *Proxy) GetProxyServers(ctx gtype.Context, ps gtype.Params) {
	data := make([]*gcfg.ProxyServerEdit, 0)
	count := len(s.cfg.ReverseProxy.Servers)
	for index := 0; index < count; index++ {
		item := &gcfg.ProxyServerEdit{}
		item.CopyFrom(s.cfg.ReverseProxy.Servers[index])
		data = append(data, item)
	}

	ctx.Success(data)
}

func (s *Proxy) GetProxyServersDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "获取服务器列表")
	function.SetNote("获取当前反向代理所有服务器信息")
	function.SetOutputDataExample([]*gcfg.ProxyServerEdit{
		{
			ProxyServerDel: gcfg.ProxyServerDel{
				Id: gtype.NewGuid(),
			},
			ProxyServerAdd: gcfg.ProxyServerAdd{
				Name:    "http",
				Disable: false,
				IP:      "",
				Port:    "80",
			},
		},
		{
			ProxyServerDel: gcfg.ProxyServerDel{
				Id: gtype.NewGuid(),
			},
			ProxyServerAdd: gcfg.ProxyServerAdd{
				Name:    "https",
				Disable: false,
				IP:      "",
				Port:    "443",
			},
		},
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) AddProxyServer(ctx gtype.Context, ps gtype.Params) {
	argument := &gcfg.ProxyServerEdit{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInput, "名称为空")
		return
	}
	if len(argument.IP) > 0 {
		addr := net.ParseIP(argument.IP)
		if addr == nil {
			ctx.Error(gtype.ErrInput, fmt.Sprintf("IP地址(%s)无效", argument.IP))
			return
		}
	}
	if len(argument.Port) < 1 {
		ctx.Error(gtype.ErrInput, "监听端口为空")
		return
	}
	port, err := strconv.ParseUint(argument.Port, 10, 16)
	if err != nil || port < 1 {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("监听端口(%s)无效", argument.Port))
		return
	}

	server := &gcfg.ProxyServer{Targets: []*gcfg.ProxyTarget{}}
	argument.CopyTo(server)
	server.Id = gtype.NewGuid()
	err = s.cfg.ReverseProxy.AddServer(server)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	err = s.saveConfig()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	ctx.Success(nil)

	go s.writeOptMessage(gtype.WSReviseProxyServerAdd, server)
}

func (s *Proxy) AddProxyServerDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "添加服务器")
	function.SetNote("添加反向代理服务器")
	function.SetInputJsonExample(&gcfg.ProxyServerAdd{
		Name:    "http",
		Disable: false,
		IP:      "",
		Port:    "80",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) DelProxyServer(ctx gtype.Context, ps gtype.Params) {
	argument := &gcfg.ProxyServer{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Id) < 1 {
		ctx.Error(gtype.ErrInput, "ID为空")
		return
	}
	err = s.cfg.ReverseProxy.DeleteServer(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	err = s.saveConfig()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	s.initRoutes()
	ctx.Success(nil)

	go s.writeOptMessage(gtype.WSReviseProxyServerDel, &gcfg.ProxyServerDel{Id: argument.Id})
}

func (s *Proxy) DelProxyServerDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "删除服务器")
	function.SetNote("删除反向代理服务器")
	function.SetInputJsonExample(&gcfg.ProxyServerDel{
		Id: gtype.NewGuid(),
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) ModifyProxyServer(ctx gtype.Context, ps gtype.Params) {
	argument := &gcfg.ProxyServerEdit{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Id) < 1 {
		ctx.Error(gtype.ErrInput, "ID为空")
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInput, "名称为空")
		return
	}
	if len(argument.IP) > 0 {
		addr := net.ParseIP(argument.IP)
		if addr == nil {
			ctx.Error(gtype.ErrInput, fmt.Sprintf("IP地址(%s)无效", argument.IP))
			return
		}
	}
	if len(argument.Port) < 1 {
		ctx.Error(gtype.ErrInput, "监听端口为空")
		return
	}
	port, err := strconv.ParseUint(argument.Port, 10, 16)
	if err != nil || port < 1 {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("监听端口(%s)无效", argument.Port))
		return
	}

	err = s.cfg.ReverseProxy.ModifyServer(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	err = s.saveConfig()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	s.initRoutes()
	ctx.Success(nil)

	go s.writeOptMessage(gtype.WSReviseProxyServerMod, argument)
}

func (s *Proxy) ModifyProxyServerDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "修改服务器")
	function.SetNote("修改反向代理服务器")
	function.SetInputJsonExample(&gcfg.ProxyServerEdit{
		ProxyServerDel: gcfg.ProxyServerDel{
			Id: gtype.NewGuid(),
		},
		ProxyServerAdd: gcfg.ProxyServerAdd{
			Name:    "http",
			Disable: false,
			IP:      "",
			Port:    "80",
		},
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) GetProxyTargets(ctx gtype.Context, ps gtype.Params) {
	argument := &gcfg.ProxyServerDel{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Id) < 1 {
		ctx.Error(gtype.ErrInput, "ID为空")
		return
	}

	server := s.cfg.ReverseProxy.GetServer(argument.Id)
	if server == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("server id '%s' not exist", argument.Id))
		return
	}

	server.InitAddrId()
	ctx.Success(server.Targets)
}

func (s *Proxy) GetProxyTargetsDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "获取目标地址列表")
	function.SetNote("获取指定反向代理所有服务器的目标地址信息")
	function.SetInputJsonExample(&gcfg.ProxyServerDel{
		Id: gtype.NewGuid(),
	})
	function.SetOutputDataExample([]*gcfg.ProxyTarget{
		{
			Id:      gtype.NewGuid(),
			Domain:  "test.com",
			IP:      "192.168.210.8",
			Port:    "8080",
			Version: 0,
			Disable: false,
			Spares: []*gcfg.ProxySpare{
				{
					IP:   "192.168.210.18",
					Port: "8080",
				},
			},
		},
		{
			Id:      gtype.NewGuid(),
			Domain:  "test.com",
			IP:      "192.168.210.17",
			Port:    "8443",
			Version: 1,
			Disable: true,
			Spares: []*gcfg.ProxySpare{
				{
					IP:   "192.168.210.27",
					Port: "8443",
				},
			},
		},
	})
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) AddProxyTarget(ctx gtype.Context, ps gtype.Params) {
	argument := &gcfg.ProxyTargetEdit{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Target.IP) < 1 {
		ctx.Error(gtype.ErrInput, "目标地址为空")
		return
	}
	if len(argument.Target.Port) < 1 {
		ctx.Error(gtype.ErrInput, "目标端口为空")
		return
	}
	port, err := strconv.ParseUint(argument.Target.Port, 10, 16)
	if err != nil || port < 1 {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("目标端口(%s)无效", argument.Target.Port))
		return
	}
	c := len(argument.Target.Spares)
	if c > 0 {
		for i := 0; i < c; i++ {
			spare := argument.Target.Spares[i]
			if spare == nil {
				ctx.Error(gtype.ErrInput, "备用目标项目为空")
				return
			}
			if len(spare.IP) < 1 {
				ctx.Error(gtype.ErrInput, "备用目标地址为空")
				return
			}
			if len(spare.Port) < 1 {
				ctx.Error(gtype.ErrInput, "备用目标端口为空")
				return
			}
		}
	}

	if len(argument.ServerId) < 1 {
		ctx.Error(gtype.ErrInput, "服务器标识ID为空")
		return
	}
	server := s.cfg.ReverseProxy.GetServer(argument.ServerId)
	if server == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("server id '%s' not exist", argument.ServerId))
		return
	}

	argument.Target.Id = gtype.NewGuid()
	err = server.AddTarget(&argument.Target)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	err = s.saveConfig()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	s.initRoutes()
	ctx.Success(nil)

	go s.writeOptMessage(gtype.WSReviseProxyTargetAdd, argument)
}

func (s *Proxy) AddProxyTargetDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "添加目标地址")
	function.SetNote("添加反向代理服务器的目标地址")
	function.SetRemark("标识ID(target.id)不需要指定")
	function.SetInputJsonExample(&gcfg.ProxyTargetEdit{
		ServerId: gtype.NewGuid(),
		Target: gcfg.ProxyTarget{
			Domain:  "test.com",
			IP:      "192.168.210.8",
			Port:    "8080",
			Version: 0,
			Disable: false,
			Spares: []*gcfg.ProxySpare{
				{
					IP:   "192.168.210.18",
					Port: "8080",
				},
			},
		},
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) DelProxyTarget(ctx gtype.Context, ps gtype.Params) {
	argument := &gcfg.ProxyTargetDel{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.ServerId) < 1 {
		ctx.Error(gtype.ErrInput, "服务器标识ID为空")
		return
	}
	if len(argument.TargetId) < 1 {
		ctx.Error(gtype.ErrInput, "目标地址标识ID为空")
		return
	}
	server := s.cfg.ReverseProxy.GetServer(argument.ServerId)
	if server == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("server id '%s' not exist", argument.ServerId))
		return
	}

	err = server.DeleteTarget(argument.TargetId)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	err = s.saveConfig()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	s.initRoutes()
	ctx.Success(nil)

	go s.writeOptMessage(gtype.WSReviseProxyTargetDel, argument)
}

func (s *Proxy) DelProxyTargetDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "删除目标地址")
	function.SetNote("删除反向代理服务器的目标地址")
	function.SetInputJsonExample(&gcfg.ProxyTargetDel{
		ServerId: gtype.NewGuid(),
		TargetId: gtype.NewGuid(),
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) ModifyProxyTarget(ctx gtype.Context, ps gtype.Params) {
	argument := &gcfg.ProxyTargetEdit{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.ServerId) < 1 {
		ctx.Error(gtype.ErrInput, "服务器标识ID为空")
		return
	}
	if len(argument.Target.Id) < 1 {
		ctx.Error(gtype.ErrInput, "目标地址标识ID为空")
		return
	}
	if len(argument.Target.IP) < 1 {
		ctx.Error(gtype.ErrInput, "目标地址为空")
		return
	}
	if len(argument.Target.Port) < 1 {
		ctx.Error(gtype.ErrInput, "目标端口为空")
		return
	}
	c := len(argument.Target.Spares)
	if c > 0 {
		for i := 0; i < c; i++ {
			spare := argument.Target.Spares[i]
			if spare == nil {
				ctx.Error(gtype.ErrInput, "备用目标项目为空")
				return
			}
			if len(spare.IP) < 1 {
				ctx.Error(gtype.ErrInput, "备用目标地址为空")
				return
			}
			if len(spare.Port) < 1 {
				ctx.Error(gtype.ErrInput, "备用目标端口为空")
				return
			}
		}
	}

	port, err := strconv.ParseUint(argument.Target.Port, 10, 16)
	if err != nil || port < 1 {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("目标端口(%s)无效", argument.Target.Port))
		return
	}
	server := s.cfg.ReverseProxy.GetServer(argument.ServerId)
	if server == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("server id '%s' not exist", argument.ServerId))
		return
	}

	target, err := server.ModifyTarget(&argument.Target)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	err = s.saveConfig()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	target.InitAddrId(argument.ServerId)
	addr := s.proxyTargets.GetItem(target.AddrId)
	if addr != nil {
		target.Alive = addr.IsAlive()
		target.ConnCount = addr.Count()
	} else {
		target.Alive = false
		target.ConnCount = 0
	}
	for i := 0; i < len(target.Spares); i++ {
		spare := target.Spares[i]
		if spare == nil {
			continue
		}
		addr = s.proxyTargets.GetItem(spare.AddrId)
		if addr != nil {
			spare.Alive = addr.IsAlive()
			spare.ConnCount = addr.Count()
		} else {
			spare.Alive = false
			spare.ConnCount = 0
		}
	}

	s.initRoutes()
	ctx.Success(nil)

	go s.writeOptMessage(gtype.WSReviseProxyTargetMod, &gcfg.ProxyTargetEdit{
		ServerId: argument.ServerId,
		Target:   *target,
	})
}

func (s *Proxy) ModifyProxyTargetDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "修改目标地址")
	function.SetNote("修改反向代理服务器的目标地址")
	function.SetInputJsonExample(&gcfg.ProxyTargetEdit{
		ServerId: gtype.NewGuid(),
		Target: gcfg.ProxyTarget{
			Id:      gtype.NewGuid(),
			Domain:  "test.com",
			IP:      "192.168.210.8",
			Port:    "8080",
			Version: 0,
			Disable: false,
			Spares: []*gcfg.ProxySpare{
				{
					IP:   "192.168.210.18",
					Port: "8080",
				},
			},
		},
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) GetProxyServiceSetting(ctx gtype.Context, ps gtype.Params) {
	data := &gmodel.ProxyServiceSetting{
		Disable: s.cfg.ReverseProxy.Disable,
	}

	ctx.Success(data)
}

func (s *Proxy) GetProxyServiceSettingDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "获取服务设置")
	function.SetNote("获取反向代理服务设置")
	function.SetOutputDataExample(&gmodel.ProxyServiceSetting{
		Disable: false,
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) SetProxyServiceSetting(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ProxyServiceSetting{
		Disable: s.cfg.ReverseProxy.Disable,
	}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if argument.Disable == s.cfg.ReverseProxy.Disable {
		ctx.Success(argument)
		return
	}

	s.cfg.ReverseProxy.Disable = argument.Disable
	err = s.saveConfig()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	if argument.Disable {
		s.proxyServer.Stop()
	}

	ctx.Success(argument)
}

func (s *Proxy) SetProxyServiceSettingDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "修改服务设置")
	function.SetNote("修改反向代理服务设置")
	function.SetInputJsonExample(&gmodel.ProxyServiceSetting{
		Disable: false,
	})
	function.SetOutputDataExample(&gmodel.ProxyServiceSetting{
		Disable: false,
	})
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) StartProxyService(ctx gtype.Context, ps gtype.Params) {
	if s.cfg.ReverseProxy.Disable {
		ctx.Error(gtype.ErrInternal, "服务已禁用")
		return
	}

	err := s.proxyServer.Start()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	ctx.Success(nil)
}

func (s *Proxy) StartProxyServiceDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "启动服务")
	function.SetNote("启动反向代理服务")
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) StopProxyService(ctx gtype.Context, ps gtype.Params) {
	if s.cfg.ReverseProxy.Disable {
		ctx.Error(gtype.ErrInternal, "服务已禁用")
		return
	}

	err := s.proxyServer.Stop()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	ctx.Success(nil)
}

func (s *Proxy) StopProxyServiceDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "停止服务")
	function.SetNote("停止反向代理服务")
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) RestartProxyService(ctx gtype.Context, ps gtype.Params) {
	if s.cfg.ReverseProxy.Disable {
		ctx.Error(gtype.ErrInternal, "服务已禁用")
		return
	}

	err := s.proxyServer.Restart()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	ctx.Success(nil)
}

func (s *Proxy) RestartProxyServiceDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "重启服务")
	function.SetNote("重启反向代理服务")
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) GetProxyServiceStatus(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(s.proxyServer.Result())
}

func (s *Proxy) GetProxyServiceStatusDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	now := gtype.DateTime(time.Now())
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "获取服务状态")
	function.SetNote("获取反向代理服务状态")
	function.SetOutputDataExample(&gproxy.Result{
		Status:    gproxy.StatusRunning,
		StartTime: &now,
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) GetProxyLinks(ctx gtype.Context, ps gtype.Params) {
	argument := &gproxy.LinkFilter{}
	ctx.GetJson(argument)
	data := s.proxyLinks.Lst(argument)

	ctx.Success(data)
}

func (s *Proxy) GetProxyLinksDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "获取连接列表")
	function.SetNote("获取当前反向代理转发连接信息")
	function.SetInputJsonExample(&gproxy.LinkFilter{})
	function.SetOutputDataExample([]*gproxy.Link{
		{
			Id:         gtype.NewGuid(),
			Time:       gtype.DateTime(time.Now()),
			ListenAddr: ":80",
			Domain:     "test.com",
			SourceAddr: "10.3.2.18:25312",
			TargetAddr: "192.168.1.6:8080",
		},
		{
			Id:         gtype.NewGuid(),
			Time:       gtype.DateTime(time.Now()),
			ListenAddr: ":443",
			Domain:     "test.com.cn",
			SourceAddr: "10.7.32.26:53127",
			TargetAddr: "192.168.1.86:8443",
		},
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) proxyCatalog(doc gtype.Doc) gtype.Catalog {
	return s.createCatalog(doc, "端口转发", "反向代理")
}

func (s *Proxy) saveConfig() error {
	cfg, err := s.cfg.Load()
	if err != nil {
		return err
	}
	s.cfg.ReverseProxy.CopyTo(&cfg.ReverseProxy)

	return s.cfg.Save(cfg)
}

func (s *Proxy) initRoutes() {
	s.proxyServer.Routes = make([]gproxy.Route, 0)
	s.proxyTargets.items = make(map[string]ProxyTargetItem)

	if s.cfg == nil {
		return
	}
	serverCount := len(s.cfg.ReverseProxy.Servers)
	for serverIndex := 0; serverIndex < serverCount; serverIndex++ {
		server := s.cfg.ReverseProxy.Servers[serverIndex]
		if server == nil {
			continue
		}
		server.InitAddrId()
		if server.Disable {
			continue
		}

		targetCount := len(server.Targets)
		for targetIndex := 0; targetIndex < targetCount; targetIndex++ {
			target := server.Targets[targetIndex]
			if target == nil {
				continue
			}
			if target.Disable {
				continue
			}

			s.proxyServer.Routes = append(s.proxyServer.Routes, gproxy.Route{
				SourceId:     server.Id,
				TargetId:     target.Id,
				IsTls:        server.TLS,
				Address:      fmt.Sprintf("%s:%s", server.IP, server.Port),
				Domain:       target.Domain,
				Path:         target.Path,
				Target:       fmt.Sprintf("%s:%s", target.IP, target.Port),
				Version:      target.Version,
				SpareTargets: target.SpareTargets(),
			})

			target.SetSourceId(server.Id)
			s.proxyTargets.AddItem(target.AddrId, target)
			spareCount := len(target.Spares)
			for spareIndex := 0; spareIndex < spareCount; spareIndex++ {
				spare := target.Spares[spareIndex]
				if spare == nil {
					continue
				}
				spare.SetSourceId(server.Id)
				spare.SetTargetId(target.Id)
				s.proxyTargets.AddItem(spare.AddrId, spare)
			}
		}
	}
}

func (s *Proxy) onProxyServerStatusChanged(status gproxy.Status) {
	if status != gproxy.StatusRunning {
		s.proxyTargets.Stop()
	}
	s.writeOptMessage(gtype.WSReviseProxyServiceStatus, s.proxyServer.Result())
}

func (s *Proxy) onProxyConnected(link gproxy.Link) {
	s.proxyLinks.Add(&link)
	s.writeOptMessage(gtype.WSReviseProxyConnectionOpen, link)
}

func (s *Proxy) onProxyDisconnected(link gproxy.Link) {
	s.proxyLinks.Del(link.Id)
	s.writeOptMessage(gtype.WSReviseProxyConnectionShut, link)
}

func (s *Proxy) onTargetAliveChanged(addr *gproxy.TargetAddressItem) {
	if addr == nil {
		return
	}

	item := s.proxyTargets.GetItem(addr.AddrId)
	if item == nil {
		return
	}

	item.SetAlive(addr.IstAlive())
	s.writeOptMessage(gtype.WSReviseProxyTargetStatusChanged, &ProxyTargetEntry{
		SourceId: item.SourceId(),
		TargetId: item.TargetId(),
		AddrId:   addr.AddrId,
		Alive:    item.IsAlive(),
		Count:    item.Count(),
	})
}

func (s *Proxy) onTargetConnCountChanged(addr *gproxy.TargetAddressItem, increase bool) {
	if addr == nil {
		return
	}

	item := s.proxyTargets.GetItem(addr.AddrId)
	if item == nil {
		return
	}
	if increase {
		item.IncreaseCount()
	} else {
		item.DecreaseCount()
	}
	s.writeOptMessage(gtype.WSReviseProxyTargetStatusChanged, &ProxyTargetEntry{
		SourceId: item.SourceId(),
		TargetId: item.TargetId(),
		AddrId:   addr.AddrId,
		Alive:    item.IsAlive(),
		Count:    item.Count(),
	})
}
