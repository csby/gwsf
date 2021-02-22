package gopt

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gopt/controller"
	"github.com/csby/gwsf/gtype"
	"net/http"
)

type Handler interface {
	Init(router gtype.Router, api func(path *gtype.Path, preHandle gtype.HttpHandle, wsc gtype.SocketChannelCollection))
}

func NewHandler(log gtype.Log, cfg *gcfg.Config, webPrefix, apiPrefix, docWebPrefix string) Handler {
	instance := &innerHandler{cfg: cfg}
	instance.SetLog(log)
	instance.svcMgr = &SvcUpdMgr{}

	instance.webPath = &gtype.Path{Prefix: webPrefix}
	instance.apiPath = &gtype.Path{Prefix: apiPrefix, DefaultTokenUI: gtype.TokenUIForAccountPassword}
	instance.docWebPrefix = docWebPrefix

	instance.wsc = gtype.NewSocketChannelCollection()
	tokenExpiredMinutes := int64(0)
	if cfg != nil {
		tokenExpiredMinutes = cfg.Site.Opt.Api.Token.Expiration
	}
	instance.dbToken = gtype.NewTokenDatabase(tokenExpiredMinutes, "opt")

	return instance
}

type innerHandler struct {
	gtype.Base

	cfg     *gcfg.Config
	wsc     gtype.SocketChannelCollection
	dbToken gtype.TokenDatabase
	svcMgr  gtype.SvcUpdMgr

	webPath      *gtype.Path
	apiPath      *gtype.Path
	docWebPrefix string

	auth      *controller.Auth
	user      *controller.User
	site      *controller.Site
	monitor   *controller.Monitor
	service   *controller.Service
	websocket *controller.Websocket
}

func (s *innerHandler) Init(router gtype.Router, api func(path *gtype.Path, preHandle gtype.HttpHandle, wsc gtype.SocketChannelCollection)) {

	tokenChecker := s.mapApi(router, s.apiPath)

	if s.cfg != nil {
		if len(s.cfg.Site.Opt.Path) > 0 {
			s.mapSite(router, s.cfg.Site.Opt.Path)
		}
	}

	if api != nil {
		api(s.apiPath, tokenChecker, s.wsc)
	}
}

func (s *innerHandler) mapApi(router gtype.Router, path *gtype.Path) gtype.HttpHandle {
	s.auth = controller.NewAuth(s.GetLog(), s.cfg, s.dbToken, s.wsc)
	s.user = controller.NewUser(s.GetLog(), s.cfg, s.dbToken, s.wsc)
	s.site = controller.NewSite(s.GetLog(), s.cfg, s.dbToken, s.wsc, s.docWebPrefix, s.webPath.Prefix)
	s.monitor = controller.NewMonitor(s.GetLog(), s.cfg)
	s.service = controller.NewService(s.GetLog(), s.cfg, s.svcMgr)
	s.websocket = controller.NewWebsocket(s.GetLog(), s.cfg, s.dbToken, s.wsc)

	s.apiPath.DefaultTokenCreate = s.auth.CreateTokenForAccountPassword
	tokenChecker := s.auth.CheckToken
	// 获取验证码
	router.POST(path.Uri("/captcha").SetTokenUI(nil).SetTokenCreate(nil), nil,
		s.auth.GetCaptcha, s.auth.GetCaptchaDoc)
	// 用户登陆
	router.POST(path.Uri("/login").SetTokenUI(nil).SetTokenCreate(nil), nil,
		s.auth.Login, s.auth.LoginDoc)
	// 注销登陆
	router.POST(path.Uri("/logout"), tokenChecker,
		s.auth.Logout, s.auth.LogoutDoc)

	// 获取登录账号
	router.POST(path.Uri("/login/account"), tokenChecker,
		s.user.GetLoginAccount, s.user.GetLoginAccountDoc)
	// 获取在线用户
	router.POST(path.Uri("/online/users"), tokenChecker,
		s.user.GetOnlineUsers, s.user.GetOnlineUsersDoc)

	// 系统信息
	router.POST(path.Uri("/monitor/host"), tokenChecker,
		s.monitor.GetHost, s.monitor.GetHostDoc)
	router.POST(path.Uri("/monitor/network/interfaces"), tokenChecker,
		s.monitor.GetNetworkInterfaces, s.monitor.GetNetworkInterfacesDoc)
	router.POST(path.Uri("/monitor/network/listen/ports"), tokenChecker,
		s.monitor.GetNetworkListenPorts, s.monitor.GetNetworkListenPortsDoc)

	// 后台服务
	router.POST(path.Uri("/service/info"), tokenChecker,
		s.service.Info, s.service.InfoDoc)
	router.POST(path.Uri("/service/restart/enable"), tokenChecker,
		s.service.CanRestart, s.service.CanRestartDoc)
	router.POST(path.Uri("/service/restart"), tokenChecker,
		s.service.Restart, s.service.RestartDoc)
	router.POST(path.Uri("/service/update/enable"), tokenChecker,
		s.service.CanUpdate, s.service.CanUpdateDoc)
	router.POST(path.Uri("/service/update"), tokenChecker,
		s.service.Update, s.service.UpdateDoc)

	// 网站管理
	router.POST(path.Uri("/site/root/file/list"), tokenChecker,
		s.site.GetRootFiles, s.site.GetRootFilesDoc)
	router.POST(path.Uri("/site/root/file/upload"), tokenChecker,
		s.site.UploadRootFile, s.site.UploadRootFileDoc)
	router.POST(path.Uri("/site/root/file/delete"), tokenChecker,
		s.site.DeleteRootFile, s.site.DeleteRootFileDoc)
	router.POST(path.Uri("/site/app/list"), tokenChecker,
		s.site.GetApps, s.site.GetAppsDoc)
	router.POST(path.Uri("/site/app/info"), tokenChecker,
		s.site.GetAppInfo, s.site.GetAppInfoDoc)
	router.POST(path.Uri("/site/app/upload"), tokenChecker,
		s.site.UploadApp, s.site.UploadAppDoc)

	// 通知推送
	router.GET(path.Uri("/websocket/notify").SetTokenPlace(gtype.TokenPlaceQuery).SetIsWebsocket(true),
		tokenChecker, s.websocket.Notify, s.websocket.NotifyDoc)

	return tokenChecker
}

func (s *innerHandler) mapSite(router gtype.Router, root string) {
	router.ServeFiles(s.webPath.Uri("/*filepath"), nil, http.Dir(root), nil)
}
