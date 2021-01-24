package gdoc

import (
	"github.com/csby/gwsf/gtype"
	"net/http"
)

const (
	WebPath = "/doc"
	ApiPath = "/doc.api"
)

const (
	ApiPathInformation    = "/info"
	ApiPathCatalogTree    = "/catalog/tree"
	ApiPathFunctionDetail = "/function/:id"
	ApiPathTokenUI        = "/token/ui/:id"
	ApiPathTokenCreate    = "/token/create/:id"
)

// rootPath: site path in location
// webPrefix: document site prefix path, http(s)://ip/[webPrefix]/*
// apiPrefix: document api prefix path, http(s)://ip/[apiPrefix]/*
func NewHandler(rootPath, webPrefix, apiPrefix string) Handler {
	return &handler{
		rootPath:   rootPath,
		sitePrefix: webPrefix,
		apiPrefix:  apiPrefix,
	}
}

type Handler interface {
	Init(router gtype.Router, info *gtype.ServerInfo)
}

type handler struct {
	rootPath   string
	sitePrefix string
	apiPrefix  string
}

func (s *handler) Init(router gtype.Router, info *gtype.ServerInfo) {
	// site
	sitePath := gtype.Path{Prefix: s.sitePrefix}
	router.ServeFiles(sitePath.Uri("/*filepath"), nil, http.Dir(s.rootPath), nil)

	// api
	apiPath := gtype.Path{Prefix: s.apiPrefix}
	ctrl := &controller{doc: router.Document()}
	if info != nil {
		ctrl.info.Name = info.Name
		ctrl.info.Version = info.Version
	}

	// 获取服务信息
	router.POST(apiPath.Uri(ApiPathInformation), nil, ctrl.GetInformation, nil)

	// 获取接口目录信息
	router.POST(apiPath.Uri(ApiPathCatalogTree), nil, ctrl.GetCatalogTree, nil)

	// 获取接口定义信息
	router.POST(apiPath.Uri(ApiPathFunctionDetail), nil, ctrl.GetFunctionDetail, nil)

	// 获取创建凭证的输入项目
	router.POST(apiPath.Uri(ApiPathTokenUI), nil, ctrl.GetTokenUI, nil)

	// 创建凭证
	router.POST(apiPath.Uri(ApiPathTokenCreate), nil, ctrl.CreateToken, nil)
}
