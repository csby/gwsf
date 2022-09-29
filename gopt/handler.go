package gopt

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gopt/controller"
	"github.com/csby/gwsf/gtype"
	"net/http"
)

type Handler interface {
	Init(router gtype.Router,
		setup func(opt gtype.Option),
		api func(path *gtype.Path, preHandle gtype.HttpHandle, opt gtype.Opt))
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

	preHandle gtype.HttpHandle
	isCloud   bool
	isNode    bool

	webPath      *gtype.Path
	apiPath      *gtype.Path
	docWebPrefix string

	auth      *controller.Auth
	role      *controller.Role
	user      *controller.User
	site      *controller.Site
	monitor   *controller.Monitor
	service   *controller.Service
	update    *controller.Update
	database  *controller.Database
	websocket *controller.Websocket
	proxy     *controller.Proxy
}

func (s *innerHandler) Init(router gtype.Router,
	setup func(opt gtype.Option),
	api func(path *gtype.Path, preHandle gtype.HttpHandle, opt gtype.Opt)) {
	if setup != nil {
		setup(s)
	}

	tokenChecker := s.mapApi(router, s.apiPath)

	if s.cfg != nil {
		if len(s.cfg.Site.Opt.Path) > 0 {
			s.mapSite(router, s.cfg.Site.Opt.Path)
		}
	}

	if api != nil {
		api(s.apiPath, tokenChecker, s.websocket)
	}
}

func (s *innerHandler) SetTokenChecker(v gtype.HttpHandle) {
	s.preHandle = v
}

func (s *innerHandler) SetCloud(v bool) {
	s.isCloud = v
}

func (s *innerHandler) SetNode(v bool) {
	s.isNode = v
}

func (s *innerHandler) mapApi(router gtype.Router, path *gtype.Path) gtype.HttpHandle {
	s.auth = controller.NewAuth(s.GetLog(), s.cfg, s.dbToken, s.wsc)
	s.user = controller.NewUser(s.GetLog(), s.cfg, s.dbToken, s.wsc)
	s.site = controller.NewSite(s.GetLog(), s.cfg, s.dbToken, s.wsc, s.docWebPrefix, s.webPath.Prefix)
	s.role = controller.NewRole(s.GetLog(), s.cfg, s.isCloud, s.isNode)
	s.monitor = controller.NewMonitor(s.GetLog(), s.cfg, s.wsc)
	s.service = controller.NewService(s.GetLog(), s.cfg, s.svcMgr, s.wsc)
	s.update = controller.NewUpdate(s.GetLog(), s.cfg, s.svcMgr)
	s.database = controller.NewDatabase(s.GetLog(), s.cfg)
	s.websocket = controller.NewWebsocket(s.GetLog(), s.cfg, s.dbToken, s.wsc)
	s.proxy = controller.NewProxy(s.GetLog(), s.cfg, s.wsc)

	if s.cfg != nil {
		s.auth.AccountVerification = s.cfg.Site.Opt.AccountVerification
	}

	s.apiPath.DefaultTokenCreate = s.auth.CreateTokenForAccountPassword
	tokenChecker := s.auth.CheckToken
	if s.preHandle != nil {
		tokenChecker = s.preHandle
	}
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
	// 获取本地用户列表
	router.POST(path.Uri("/user/local/list"), tokenChecker,
		s.user.GetList, s.user.GetListDoc)
	// 新建本地用户
	router.POST(path.Uri("/user/local/create"), tokenChecker,
		s.user.Create, s.user.CreateDoc)
	// 删除本地用户
	router.POST(path.Uri("/user/local/delete"), tokenChecker,
		s.user.Delete, s.user.DeleteDoc)
	// 修改本地用户
	router.POST(path.Uri("/user/local/modify"), tokenChecker,
		s.user.Modify, s.user.ModifyDoc)
	// 重置本地用户密码
	router.POST(path.Uri("/user/local/password/reset"), tokenChecker,
		s.user.ResetPassword, s.user.ResetPasswordDoc)
	// 修改本地用户密码
	router.POST(path.Uri("/user/local/password/change"), tokenChecker,
		s.user.ChangePassword, s.user.ChangePasswordDoc)
	// 获取LDAP设置
	router.POST(path.Uri("/user/ldap/get"), tokenChecker,
		s.auth.GetLdap, s.auth.GetLdapDoc)
	//  修改LDAP设置
	router.POST(path.Uri("/user/ldap/set"), tokenChecker,
		s.auth.SetLdap, s.auth.SetLdapDoc)

	// 系统角色
	router.POST(path.Uri("/sys/role/server"), tokenChecker,
		s.role.GetServerRole, s.role.GetServerRoleDoc)

	// 系统资源
	router.POST(path.Uri("/monitor/host"), tokenChecker,
		s.monitor.GetHost, s.monitor.GetHostDoc)
	router.POST(path.Uri("/monitor/network/interfaces"), tokenChecker,
		s.monitor.GetNetworkInterfaces, s.monitor.GetNetworkInterfacesDoc)
	router.POST(path.Uri("/monitor/network/throughput/list"), tokenChecker,
		s.monitor.GetNetworkThroughput, s.monitor.GetNetworkThroughputDoc)
	router.POST(path.Uri("/monitor/network/listen/ports"), tokenChecker,
		s.monitor.GetNetworkListenPorts, s.monitor.GetNetworkListenPortsDoc)
	router.POST(path.Uri("/monitor/cpu/usage/list"), tokenChecker,
		s.monitor.GetCpuUsage, s.monitor.GetCpuUsageDoc)
	router.POST(path.Uri("/monitor/mem/usage/list"), tokenChecker,
		s.monitor.GetMemoryUsage, s.monitor.GetMemoryUsageDoc)
	router.POST(path.Uri("/monitor/disk/usage/list"), tokenChecker,
		s.monitor.GetDiskPartitionUsages, s.monitor.GetDiskPartitionUsagesDoc)
	router.POST(path.Uri("/monitor/process/info"), tokenChecker,
		s.monitor.GetProcessInfo, s.monitor.GetProcessInfoDoc)

	// 后台服务
	router.POST(path.Uri("/service/version").SetTokenUI(nil).SetTokenCreate(nil), nil,
		s.service.Version, s.service.VersionDoc)
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

	// 更新管理
	router.POST(path.Uri("/update/enable"), tokenChecker,
		s.update.Enable, s.update.EnableDoc)
	router.POST(path.Uri("/update/info"), tokenChecker,
		s.update.Info, s.update.InfoDoc)
	router.POST(path.Uri("/update/restart/enable"), tokenChecker,
		s.update.CanRestart, s.update.CanRestartDoc)
	router.POST(path.Uri("/update/restart"), tokenChecker,
		s.update.Restart, s.update.RestartDoc)
	router.POST(path.Uri("/update/upload/enable"), tokenChecker,
		s.update.CanUpdate, s.update.CanUpdateDoc)
	router.POST(path.Uri("/update/upload"), tokenChecker,
		s.update.Update, s.update.UpdateDoc)

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

	// 数据库
	router.POST(path.Uri("/db/mssql/instance/list"), tokenChecker,
		s.database.GetSqlServerInstances, s.database.GetSqlServerInstancesDoc)

	// 反向代理-服务
	router.POST(path.Uri("/proxy/service/setting/get"), tokenChecker,
		s.proxy.GetProxyServiceSetting, s.proxy.GetProxyServiceSettingDoc)
	router.POST(path.Uri("/proxy/service/setting/set"), tokenChecker,
		s.proxy.SetProxyServiceSetting, s.proxy.SetProxyServiceSettingDoc)
	router.POST(path.Uri("/proxy/service/status"), tokenChecker,
		s.proxy.GetProxyServiceStatus, s.proxy.GetProxyServiceStatusDoc)
	router.POST(path.Uri("/proxy/service/start"), tokenChecker,
		s.proxy.StartProxyService, s.proxy.StartProxyServiceDoc)
	router.POST(path.Uri("/proxy/service/stop"), tokenChecker,
		s.proxy.StopProxyService, s.proxy.StopProxyServiceDoc)
	router.POST(path.Uri("/proxy/service/restart"), tokenChecker,
		s.proxy.RestartProxyService, s.proxy.RestartProxyServiceDoc)

	// 反向代理-连接
	router.POST(path.Uri("/proxy/conn/list"), tokenChecker,
		s.proxy.GetProxyLinks, s.proxy.GetProxyLinksDoc)

	// 反向代理-端口
	router.POST(path.Uri("/proxy/server/list"), tokenChecker,
		s.proxy.GetProxyServers, s.proxy.GetProxyServersDoc)
	router.POST(path.Uri("/proxy/server/add"), tokenChecker,
		s.proxy.AddProxyServer, s.proxy.AddProxyServerDoc)
	router.POST(path.Uri("/proxy/server/del"), tokenChecker,
		s.proxy.DelProxyServer, s.proxy.DelProxyServerDoc)
	router.POST(path.Uri("/proxy/server/mod"), tokenChecker,
		s.proxy.ModifyProxyServer, s.proxy.ModifyProxyServerDoc)

	// 反向代理-目标
	router.POST(path.Uri("/proxy/target/list"), tokenChecker,
		s.proxy.GetProxyTargets, s.proxy.GetProxyTargetsDoc)
	router.POST(path.Uri("/proxy/target/add"), tokenChecker,
		s.proxy.AddProxyTarget, s.proxy.AddProxyTargetDoc)
	router.POST(path.Uri("/proxy/target/del"), tokenChecker,
		s.proxy.DelProxyTarget, s.proxy.DelProxyTargetDoc)
	router.POST(path.Uri("/proxy/target/mod"), tokenChecker,
		s.proxy.ModifyProxyTarget, s.proxy.ModifyProxyTargetDoc)

	// 系统服务-tomcat
	router.POST(path.Uri("/svc/tomcat/svc/list"), tokenChecker,
		s.service.GetTomcats, s.service.GetTomcatsDoc)
	router.POST(path.Uri("/svc/tomcat/svc/start"), tokenChecker,
		s.service.StartTomcat, s.service.StartTomcatDoc)
	router.POST(path.Uri("/svc/tomcat/svc/stop"), tokenChecker,
		s.service.StopTomcat, s.service.StopTomcatDoc)
	router.POST(path.Uri("/svc/tomcat/svc/restart"), tokenChecker,
		s.service.RestartTomcat, s.service.RestartTomcatDoc)

	router.POST(path.Uri("/svc/tomcat/app/list"), tokenChecker,
		s.service.GetTomcatApps, s.service.GetTomcatAppsDoc)
	router.GET(path.Uri("/svc/tomcat/app/download/:name/:app"), tokenChecker,
		s.service.DownloadTomcatApp, s.service.DownloadTomcatAppDoc)
	router.POST(path.Uri("/svc/tomcat/app/mod"), tokenChecker,
		s.service.ModTomcatApp, s.service.ModTomcatAppDoc)
	router.POST(path.Uri("/svc/tomcat/app/del"), tokenChecker,
		s.service.DelTomcatApp, s.service.DelTomcatAppDoc)
	router.POST(path.Uri("/svc/tomcat/app/detail"), tokenChecker,
		s.service.GetTomcatAppDetail, s.service.GetTomcatDetailDoc)

	router.POST(path.Uri("/svc/tomcat/cfg/tree"), tokenChecker,
		s.service.GetTomcatConfigs, s.service.GetTomcatConfigsDoc)
	router.GET(path.Uri("/svc/tomcat/cfg/file/content/:name/:path"), tokenChecker,
		s.service.ViewTomcatConfigFile, s.service.ViewTomcatConfigFileDoc)
	router.GET(path.Uri("/svc/tomcat/cfg/file/download/:name/:path"), tokenChecker,
		s.service.DownloadTomcatConfigFile, s.service.DownloadTomcatConfigFileDoc)
	router.POST(path.Uri("/svc/tomcat/cfg/folder/add"), tokenChecker,
		s.service.CreateTomcatConfigFolder, s.service.CreateTomcatConfigFolderDoc)
	router.POST(path.Uri("/svc/tomcat/cfg/mod"), tokenChecker,
		s.service.ModTomcatConfig, s.service.ModTomcatConfigDoc)
	router.POST(path.Uri("/svc/tomcat/cfg/del"), tokenChecker,
		s.service.DeleteTomcatConfig, s.service.DeleteTomcatConfigDoc)

	router.POST(path.Uri("/svc/tomcat/log/tree"), tokenChecker,
		s.service.GetTomcatLogs, s.service.GetTomcatLogsDoc)
	router.GET(path.Uri("/svc/tomcat/log/file/content/:name/:path"), tokenChecker,
		s.service.ViewTomcatLogFile, s.service.ViewTomcatLogFileDoc)
	router.GET(path.Uri("/svc/tomcat/log/file/download/:name/:path"), tokenChecker,
		s.service.DownloadTomcatLogFile, s.service.DownloadTomcatLogFileDoc)
	router.POST(path.Uri("/svc/tomcat/log/del"), tokenChecker,
		s.service.DeleteTomcatLog, s.service.DeleteTomcatLogDoc)

	// 系统服务-nginx
	router.POST(path.Uri("/svc/nginx/svc/list"), tokenChecker,
		s.service.GetNginxes, s.service.GetNginxesDoc)
	router.POST(path.Uri("/svc/nginx/svc/start"), tokenChecker,
		s.service.StartNginx, s.service.StartNginxDoc)
	router.POST(path.Uri("/svc/nginx/svc/stop"), tokenChecker,
		s.service.StopNginx, s.service.StopNginxDoc)
	router.POST(path.Uri("/svc/nginx/svc/restart"), tokenChecker,
		s.service.RestartNginx, s.service.RestartNginxDoc)
	router.POST(path.Uri("/svc/nginx/app/mod"), tokenChecker,
		s.service.ModNginxApp, s.service.ModNginxAppDoc)
	router.POST(path.Uri("/svc/nginx/app/detail"), tokenChecker,
		s.service.GetNginxAppDetail, s.service.GetNginxAppDetailDoc)

	router.POST(path.Uri("/svc/nginx/log/tree"), tokenChecker,
		s.service.GetNginxLogs, s.service.GetNginxLogsDoc)
	router.GET(path.Uri("/svc/nginx/log/file/content/:name/:path"), tokenChecker,
		s.service.ViewNginxLogFile, s.service.ViewNginxLogFileDoc)
	router.GET(path.Uri("/svc/nginx/log/file/download/:name/:path"), tokenChecker,
		s.service.DownloadNginxLogFile, s.service.DownloadNginxLogFileDoc)
	router.POST(path.Uri("/svc/nginx/log/del"), tokenChecker,
		s.service.DeleteNginxLog, s.service.DeleteNginxLogDoc)

	// 系统服务-自定义
	router.POST(path.Uri("/svc/custom/shell/info"), tokenChecker,
		s.service.GetCustomShellInfo, s.service.GetCustomShellInfoDoc)
	router.POST(path.Uri("/svc/custom/shell/update"), tokenChecker,
		s.service.UpdateCustomShell, s.service.UpdateCustomShellDoc)
	router.POST(path.Uri("/svc/custom/list"), tokenChecker,
		s.service.GetCustoms, s.service.GetCustomsDoc)
	router.POST(path.Uri("/svc/custom/add"), tokenChecker,
		s.service.AddCustom, s.service.AddCustomDoc)
	router.POST(path.Uri("/svc/custom/mod"), tokenChecker,
		s.service.ModCustom, s.service.ModCustomDoc)
	router.POST(path.Uri("/svc/custom/del"), tokenChecker,
		s.service.DelCustom, s.service.DelCustomDoc)
	router.GET(path.Uri("/svc/custom/download/:name"), tokenChecker,
		s.service.DownloadCustom, s.service.DownloadCustomDoc)
	router.POST(path.Uri("/svc/custom/install"), tokenChecker,
		s.service.InstallCustom, s.service.InstallCustomDoc)
	router.POST(path.Uri("/svc/custom/uninstall"), tokenChecker,
		s.service.UninstallCustom, s.service.UninstallCustomDoc)
	router.POST(path.Uri("/svc/custom/start"), tokenChecker,
		s.service.StartCustom, s.service.StartCustomDoc)
	router.POST(path.Uri("/svc/custom/stop"), tokenChecker,
		s.service.StopCustom, s.service.StopCustomDoc)
	router.POST(path.Uri("/svc/custom/restart"), tokenChecker,
		s.service.RestartCustom, s.service.RestartCustomDoc)
	router.POST(path.Uri("/svc/custom/app/detail"), tokenChecker,
		s.service.GetCustomDetail, s.service.GetCustomDetailDoc)

	router.POST(path.Uri("/svc/custom/log/file/list"), tokenChecker,
		s.service.GetCustomLogFiles, s.service.GetCustomLogFilesDoc)
	router.GET(path.Uri("/svc/custom/log/file/download/:path"), tokenChecker,
		s.service.DownloadCustomLogFile, s.service.DownloadCustomLogFileDoc)
	router.GET(path.Uri("/svc/custom/log/file/content/:path"), tokenChecker,
		s.service.ViewCustomLogFile, s.service.ViewCustomLogFileDoc)

	// 系统服务-其他
	router.POST(path.Uri("/svc/other/svc/list"), tokenChecker,
		s.service.GetOthers, s.service.GetOthersDoc)
	router.POST(path.Uri("/svc/other/svc/start"), tokenChecker,
		s.service.StartOther, s.service.StartOtherDoc)
	router.POST(path.Uri("/svc/other/svc/stop"), tokenChecker,
		s.service.StopOther, s.service.StopOtherDoc)
	router.POST(path.Uri("/svc/other/svc/restart"), tokenChecker,
		s.service.RestartOther, s.service.RestartOtherDoc)

	// 系统服务-文件
	router.GET(path.Uri("/svc/file/content/:path"), tokenChecker,
		s.service.ViewFile, s.service.ViewFileDoc)
	router.GET(path.Uri("/svc/file/download/:path"), tokenChecker,
		s.service.DownloadFile, s.service.DownloadFileDoc)
	router.POST(path.Uri("/svc/file/mod"), tokenChecker,
		s.service.ModFile, s.service.ModFileDoc)
	router.POST(path.Uri("/svc/file/del"), tokenChecker,
		s.service.DeleteFile, s.service.DeleteFileDoc)
	fileServers := s.service.FileServers()
	fsc := len(fileServers)
	for fsi := 0; fsi < fsc; fsi++ {
		fs := fileServers[fsi]
		if fs == nil {
			continue
		}
		if len(fs.Path) < 1 {
			continue
		}

		status := "; disabled"
		if fs.Enabled {
			status = "; enabled"

			webPath := &gtype.Path{Prefix: fs.Path}
			router.ServeFiles(webPath.Uri("/*filepath"), nil, http.Dir(fs.Root), nil)

			router.POST(webPath.Uri(""), nil,
				fs.Upload, nil)
		}

		s.LogInfo("file server path: ", fs.Path, status)
	}

	// 通知推送
	router.GET(path.Uri("/websocket/notify").SetTokenPlace(gtype.TokenPlaceQuery).SetIsWebsocket(true),
		tokenChecker, s.websocket.Notify, s.websocket.NotifyDoc)

	return tokenChecker
}

func (s *innerHandler) mapSite(router gtype.Router, root string) {
	router.ServeFiles(s.webPath.Uri("/*filepath"), nil, http.Dir(root), nil)
}
