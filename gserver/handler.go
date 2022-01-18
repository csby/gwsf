package gserver

import (
	"fmt"
	"github.com/csby/gsecurity/gcrt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gdoc"
	"github.com/csby/gwsf/gopt"
	"github.com/csby/gwsf/grouter"
	"github.com/csby/gwsf/gtype"
	"net"
	"net/http"
	"time"
)

func newHandler(log gtype.Log, cfg *gcfg.Config, hdl gtype.Handler) (*handler, error) {
	instance := &handler{handler: hdl, cfg: cfg, router: grouter.New()}
	instance.SetLog(log)

	clusterIndex := uint64(0)
	documentEnabled := false
	documentRoot := ""
	serverInfo := &gtype.ServerInfo{}
	appSiteCount := 0
	if cfg != nil {
		clusterIndex = cfg.Cluster.Index
		documentEnabled = cfg.Site.Doc.Enabled
		documentRoot = cfg.Site.Doc.Path
		serverInfo.Name = cfg.Module.Remark
		serverInfo.Version = cfg.Module.Version

		appSiteCount = len(cfg.Site.Apps)
		instance.router.NotFound = &notFound{root: cfg.Site.Root.Path}
	}

	instance.rid = gtype.NewRand(clusterIndex)
	instance.router.Doc = gdoc.NewDoc(documentEnabled)

	instance.router.Doc.OnFunctionReady(func(index int, method, path, name string) {
		if log != nil {
			log.Debug(fmt.Sprintf("api-%03d", index), ": [", method, "] ", path, " (", name, ") has been ready")
		}
	})

	otpHandler := gopt.NewHandler(log, cfg, gopt.WebPath, gopt.ApiPath, gdoc.WebPath)
	otpHandler.Init(instance.router,
		func(opt gtype.Option) {
			if hdl != nil {
				hdl.ExtendOptSetup(opt)
			}
		},
		func(path *gtype.Path, preHandle gtype.HttpHandle, wsc gtype.SocketChannelCollection) {
			if hdl != nil {
				hdl.ExtendOptApi(instance.router, path, preHandle, wsc)
			}
		})

	if hdl != nil {
		hdl.InitRouting(instance.router)
	}

	if documentEnabled {
		docHandler := gdoc.NewHandler(documentRoot, gdoc.WebPath, gdoc.ApiPath)
		docHandler.Init(instance.router, serverInfo)

		if log != nil {
			log.Info("document for api is enabled")
			log.Info("document information api path: [POST] ", gdoc.ApiPath, gdoc.ApiPathInformation)
			log.Info("document catalog api path: [POST] ", gdoc.ApiPath, gdoc.ApiPathCatalogTree)
			log.Info("document function api path: [POST] ", gdoc.ApiPath, gdoc.ApiPathFunctionDetail)
			log.Info("document token ui api path: [POST] ", gdoc.ApiPath, gdoc.ApiPathTokenUI)
			log.Info("document token create api path: [POST] ", gdoc.ApiPath, gdoc.ApiPathTokenCreate)
		}
	}

	for appSiteIndex := 0; appSiteIndex < appSiteCount; appSiteIndex++ {
		appSite := cfg.Site.Apps[appSiteIndex]
		appPath := gtype.Path{Prefix: appSite.Uri}
		instance.router.ServeFiles(appPath.Uri("/*filepath"), nil, http.Dir(appSite.Path), nil)
		log.Info(fmt.Sprintf("webapp [%d/%d] '%s' is ready: uri=%s, path=%s",
			appSiteIndex+1, appSiteCount,
			appSite.Name, appSite.Uri, appSite.Path))
	}

	return instance, nil
}

type handler struct {
	gtype.Base

	cfg     *gcfg.Config
	handler gtype.Handler

	router *grouter.Router
	rid    gtype.Rand
}

func (s *handler) ServeHTTP(w http.ResponseWriter, r *http.Request, caCrt *gcrt.Crt, serverCrt *gcrt.Pfx) {
	ctx := s.newContext(w, r)
	ctx.certificate.Ca = caCrt
	ctx.certificate.Server = serverCrt

	s.LogDebug("new request: rid=", ctx.rid,
		", rip=", ctx.rip,
		", host=", r.Host,
		", schema=", ctx.schema,
		", method=", r.Method,
		", path=", ctx.path,
		", token=", ctx.token)

	defer func(ctx *context) {
		ctx.leaveTime = time.Now()
		go s.afterRouting(ctx)
	}(ctx)

	defer func(ctx *context) {
		if err := recover(); err != nil {
			s.LogError(ctx.schema,
				" request error(rid=", ctx.rid,
				", schema=", ctx.schema,
				", path=", ctx.path,
				", rip=", ctx.rip,
				"): ", err)

			ctx.Error(gtype.ErrException, err)
		}
	}(ctx)

	ctx.path = r.URL.Path
	s.beforeRouting(ctx)
	if ctx.IsHandled() {
		return
	}

	s.router.Serve(ctx)
}

func (s *handler) beforeRouting(ctx *context) {
	if s.handler == nil {
		return
	}

	s.handler.BeforeRouting(ctx)
}

func (s *handler) afterRouting(ctx *context) {
	if s.handler == nil {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			s.LogError("afterRouting", err)
		}
	}()

	s.handler.AfterRouting(ctx)
}

func (s *handler) newContext(w http.ResponseWriter, r *http.Request) *context {
	ctx := &context{response: w, request: r, schema: "http"}
	ctx.method = r.Method
	if r.TLS != nil {
		ctx.schema = "https"
		if len(r.TLS.PeerCertificates) > 0 {
			clientCrt := &gcrt.Crt{}
			if clientCrt.FromConnectionState(r.TLS) == nil {
				ctx.certificate.Client = clientCrt
			}
		}
	}
	ctx.keys = make(map[string]interface{})
	ctx.log = false
	ctx.enterTime = time.Now()
	ctx.path = r.URL.Path
	ctx.rid = s.rid.New()
	ctx.rip, _, _ = net.SplitHostPort(r.RemoteAddr)
	ctx.token = r.Header.Get("token")
	if ctx.token == "" {
		if r.Method == "GET" {
			ctx.token = r.FormValue("token")
		}
	}
	if len(r.URL.Query()) > 0 {
		ctx.queries = make(gtype.QueryCollection, 0)
		for k, v := range r.URL.Query() {
			query := &gtype.Query{Key: k}
			if len(v) > 0 {
				query.Value = v[0]
			}
			ctx.queries = append(ctx.queries, query)
		}
	}

	return ctx
}
