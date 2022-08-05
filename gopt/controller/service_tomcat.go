package controller

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/csby/gwsf/gfile"
	"github.com/csby/gwsf/gmodel"
	"github.com/csby/gwsf/gtype"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func (s *Service) StartTomcat(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServerArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInternal, "服务名称(name)为空")
		return
	}
	info := s.cfg.Sys.Svc.GetTomcatByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}

	err = s.start(argument.Name)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	svcStatus := &gmodel.ServiceStatus{Name: argument.Name}
	svcStatus.Status, err = s.getStatus(argument.Name)
	if err == nil {
		go s.writeOptMessage(gtype.WSSvcStatusChanged, svcStatus)
	}

	ctx.Success(info)
}

func (s *Service) StartTomcatDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "启动服务")
	function.SetInputJsonExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) StopTomcat(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServerArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInternal, "服务名称(name)为空")
		return
	}
	info := s.cfg.Sys.Svc.GetTomcatByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}

	err = s.stop(argument.Name)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	svcStatus := &gmodel.ServiceStatus{Name: argument.Name}
	svcStatus.Status, err = s.getStatus(argument.Name)
	if err == nil {
		go s.writeOptMessage(gtype.WSSvcStatusChanged, svcStatus)
	}

	ctx.Success(info)
}

func (s *Service) StopTomcatDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "停止服务")
	function.SetInputJsonExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) RestartTomcat(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServerArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInternal, "服务名称(name)为空")
		return
	}
	info := s.cfg.Sys.Svc.GetTomcatByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}

	err = s.restart(argument.Name)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	svcStatus := &gmodel.ServiceStatus{Name: argument.Name}
	svcStatus.Status, err = s.getStatus(argument.Name)
	if err == nil {
		go s.writeOptMessage(gtype.WSSvcStatusChanged, svcStatus)
	}

	ctx.Success(info)
}

func (s *Service) RestartTomcatDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "重启服务")
	function.SetInputJsonExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) GetTomcats(ctx gtype.Context, ps gtype.Params) {
	results := make([]*gmodel.ServiceTomcatInfo, 0)

	if s.cfg != nil {
		items := s.cfg.Sys.Svc.Tomcats
		c := len(items)
		for i := 0; i < c; i++ {
			item := items[i]
			if item == nil {
				continue
			}
			if len(item.ServiceName) < 1 {
				continue
			}

			result := &gmodel.ServiceTomcatInfo{
				Name:        item.Name,
				ServiceName: item.ServiceName,
				WebApp:      item.WebApp,
				WebCfg:      item.WebCfg,
				WebLog:      item.WebLog,
				WebUrls:     item.WebUrls,
			}
			if len(result.Name) < 1 {
				result.Name = result.ServiceName
			}
			result.Status, _ = s.getStatus(result.ServiceName)

			results = append(results, result)
		}
	}

	ctx.Success(results)
}

func (s *Service) GetTomcatsDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "获取服务列表")
	function.SetOutputDataExample([]*gmodel.ServiceTomcatInfo{
		{
			Name:        "example",
			ServiceName: "tomcat",
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) GetTomcatApps(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServerArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInternal, "服务名称(name)为空")
		return
	}

	info := s.cfg.Sys.Svc.GetTomcatByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}

	results := make([]*gmodel.ServiceTomcatApp, 0)
	folder := info.WebApp
	if len(folder) > 0 {
		fs, fe := ioutil.ReadDir(folder)
		if fe == nil {
			for _, f := range fs {
				if f.IsDir() {
					result := &gmodel.ServiceTomcatApp{
						Name:       f.Name(),
						DeployTime: gtype.DateTime(f.ModTime()),
					}
					result.Version = s.getTomcatAppVersion(filepath.Join(folder, result.Name))
					results = append(results, result)
				}
			}
		}
	}

	ctx.Success(results)
}

func (s *Service) GetTomcatAppsDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "获取应用列表")
	function.SetInputJsonExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.SetOutputDataExample([]*gmodel.ServiceTomcatApp{
		{
			Name:       "example",
			Version:    "1.0.1.0",
			DeployTime: gtype.DateTime(time.Now()),
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) DownloadTomcatApp(ctx gtype.Context, ps gtype.Params) {
	name := ps.ByName("name")
	if len(name) < 1 {
		ctx.Error(gtype.ErrInput, "服务名称(name)为空")
		return
	}
	app := ps.ByName("app")
	if len(app) < 1 {
		ctx.Error(gtype.ErrInput, "应用名称(app)为空")
		return
	}

	info := s.cfg.Sys.Svc.GetTomcatByServiceName(name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", name))
		return
	}

	rootFolder := info.WebApp
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
		return
	}

	appFolder := filepath.Join(rootFolder, app)
	fileName := fmt.Sprintf("%s.war", app)
	ctx.Response().Header().Set("Content-Disposition", fmt.Sprint("attachment; filename=", fileName))

	s.compressFolder(ctx.Response(), appFolder, "", nil)
}

func (s *Service) DownloadTomcatAppDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "下载应用程序")
	function.SetNote("应用程序文件(.war)")
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) ModTomcatApp(ctx gtype.Context, ps gtype.Params) {
	name := strings.TrimSpace(ctx.Request().FormValue("name"))
	if len(name) < 1 {
		ctx.Error(gtype.ErrInput, "服务名称(name)为空")
		return
	}

	info := s.cfg.Sys.Svc.GetTomcatByServiceName(name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", name))
		return
	}

	rootFolder := info.WebApp
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
		return
	}

	uploadFile, head, err := ctx.Request().FormFile("file")
	if err != nil {
		ctx.Error(gtype.ErrInput, "上传文件无效: ", err)
		return
	}
	defer uploadFile.Close()

	buf := &bytes.Buffer{}
	fileSize, err := buf.ReadFrom(uploadFile)
	if err != nil {
		ctx.Error(gtype.ErrInput, "读取文件上传文件失败: ", err)
		return
	}
	if fileSize < 1 {
		ctx.Error(gtype.ErrInput, "上传的文件无效: 文件大小为0")
		return
	}

	tempFolder := filepath.Join(filepath.Dir(rootFolder), ctx.NewGuid())
	err = os.MkdirAll(tempFolder, 0777)
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("创建临时文件夹'%s'失败: ", tempFolder), err)
		return
	}
	defer os.RemoveAll(tempFolder)

	fileData := buf.Bytes()
	zipFile := &gfile.Zip{}
	err = zipFile.DecompressMemory(fileData, tempFolder)
	if err != nil {
		tarFile := &gfile.Tar{}
		err = tarFile.DecompressMemory(fileData, tempFolder)
		if err != nil {
			ctx.Error(gtype.ErrInternal, "解压文件失败: ", err)
			return
		}
	}

	app := strings.TrimSuffix(head.Filename, path.Ext(head.Filename))
	appFolder := filepath.Join(rootFolder, app)
	err = os.RemoveAll(appFolder)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "删除原应用程序失败: ", err)
		return
	}

	err = os.Rename(tempFolder, appFolder)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "重命名文件夹失败: ", err)
		return
	}

	argument := &gmodel.ServiceTomcatArgument{
		Name: name,
		App:  app,
	}

	go s.writeOptMessage(gtype.WSTomcatAppUpdated, argument)

	ctx.Success(argument)
}

func (s *Service) ModTomcatAppDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "上传应用程序")
	function.SetNote("上传应用程序打包文件(.war, .zip或.tar.gz)，并解压缩到webapps目录下，成功时返回服务名称及应用名称信息")
	function.SetRemark("使用上传文件的文件名作为应用程序名称, 如上传文件'example.war'的应用名称为'example'")
	function.AddInputHeader(true, "content-type", "内容类型", gtype.ContentTypeFormData)
	function.AddInputForm(true, "name", "tomcat服务名称", gtype.FormValueKindText, "")
	function.AddInputForm(true, "file", "应用程序打包文件(.war, .zip或.tar.gz)", gtype.FormValueKindFile, nil)
	function.SetOutputDataExample(&gmodel.ServiceTomcatArgument{
		Name: "example",
		App:  "api",
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) DelTomcatApp(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServiceTomcatArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInput, "服务名称(name)为空")
		return
	}
	if len(argument.App) < 1 {
		ctx.Error(gtype.ErrInput, " 程序名称(app)为空")
		return
	}
	info := s.cfg.Sys.Svc.GetTomcatByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}

	rootFolder := info.WebApp
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "物理根路径为空")
		return
	}

	appFolder := filepath.Join(rootFolder, argument.App)
	err = os.RemoveAll(appFolder)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "删除文件夹失败: ", err)
		return
	}

	go s.writeOptMessage(gtype.WSTomcatAppDeleted, argument)

	ctx.Success(argument)
}

func (s *Service) DelTomcatAppDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "删除应用程序")
	function.SetRemark("删除webapps目录下的应用程序")
	function.SetInputJsonExample(&gmodel.ServiceTomcatArgument{
		Name: "example",
		App:  "app",
	})
	function.SetOutputDataExample(&gmodel.ServiceTomcatArgument{
		Name: "example",
		App:  "app",
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) GetTomcatAppDetail(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServiceTomcatArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInput, "服务名称(name)为空")
		return
	}
	if len(argument.App) < 1 {
		ctx.Error(gtype.ErrInput, " 程序名称(app)为空")
		return
	}
	info := s.cfg.Sys.Svc.GetTomcatByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}

	rootFolder := info.WebApp
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "物理根路径为空")
		return
	}
	appFolder := filepath.Join(rootFolder, argument.App)

	cfg := &gmodel.FileInfo{}
	s.getFileInfos(cfg, appFolder)

	cfg.Sort()
	ctx.Success(cfg)
}

func (s *Service) GetTomcatDetailDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "获取应用程序序详细信息")
	function.SetInputJsonExample(&gmodel.ServiceTomcatArgument{
		Name: "svc",
		App:  "app",
	})
	function.SetOutputDataExample(&gmodel.FileInfo{})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) GetTomcatConfigs(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServerArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInternal, "服务名称(name)为空")
		return
	}

	info := s.cfg.Sys.Svc.GetTomcatByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}

	cfg := &gmodel.ServiceTomcatCfg{}
	s.getTomcatCfg(cfg, info.WebCfg, "")

	cfg.Sort()
	ctx.Success(cfg.Children)
}

func (s *Service) GetTomcatConfigsDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "获取应用配置列表")
	function.SetInputJsonExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.SetOutputDataExample([]*gmodel.ServiceTomcatCfg{
		{
			Name:     "example",
			Path:     base64.URLEncoding.EncodeToString([]byte("example")),
			Children: []*gmodel.ServiceTomcatCfg{},
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) ViewTomcatConfigFile(ctx gtype.Context, ps gtype.Params) {
	name := ps.ByName("name")
	if len(name) < 1 {
		ctx.Error(gtype.ErrInput, "服务名称(name)为空")
		return
	}

	pathName := ps.ByName("path")
	if len(pathName) < 1 {
		ctx.Error(gtype.ErrInput, "base64路径为空")
		return
	}

	pathData, err := base64.URLEncoding.DecodeString(pathName)
	if err != nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("路径base64(%s)无效", pathName))
		return
	}
	if len(pathData) < 1 {
		ctx.Error(gtype.ErrInput, "路径为空")
		return
	}

	info := s.cfg.Sys.Svc.GetTomcatByServiceName(name)
	if info == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("服务(%s)不存在", name))
		return
	}

	rootFolder := info.WebCfg
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "配置物理根路径为空")
		return
	}

	fullPath := filepath.Join(rootFolder, string(pathData))
	fi, fe := os.Stat(fullPath)
	if os.IsNotExist(fe) {
		ctx.Error(gtype.ErrInternal, fe)
		return
	}
	if fi.IsDir() {
		ctx.Error(gtype.ErrInternal, "指定的路径为文件夹")
		return
	}

	cfgFile, le := os.OpenFile(fullPath, os.O_RDONLY, 0666)
	if le != nil {
		ctx.Error(gtype.ErrInternal, le)
		return
	}
	defer cfgFile.Close()

	extName := strings.ToLower(path.Ext(fullPath))
	if extName == ".xml" {
		ctx.Response().Header().Set("Content-Type", "application/xml;charset=utf-8")
	} else if extName == ".json" {
		ctx.Response().Header().Set("Content-Type", "application/json;charset=utf-8")
	} else {
		ctx.Response().Header().Set("Content-Type", "text/plain;charset=utf-8")
	}

	contentLength := fi.Size()
	ctx.Response().Header().Set("Content-Length", fmt.Sprint(contentLength))

	io.Copy(ctx.Response(), cfgFile)
}

func (s *Service) ViewTomcatConfigFileDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "查看应用配置文件")
	function.SetNote("返回应用配置文本内容")
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) DownloadTomcatConfigFile(ctx gtype.Context, ps gtype.Params) {
	name := ps.ByName("name")
	if len(name) < 1 {
		ctx.Error(gtype.ErrInput, "服务名称(name)为空")
		return
	}

	pathName := ps.ByName("path")
	if len(pathName) < 1 {
		ctx.Error(gtype.ErrInput, "base64路径为空")
		return
	}

	pathData, err := base64.URLEncoding.DecodeString(pathName)
	if err != nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("路径base64(%s)无效", pathName))
		return
	}
	if len(pathData) < 1 {
		ctx.Error(gtype.ErrInput, "路径为空")
		return
	}

	info := s.cfg.Sys.Svc.GetTomcatByServiceName(name)
	if info == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("服务(%s)不存在", name))
		return
	}

	rootFolder := info.WebCfg
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "配置物理根路径为空")
		return
	}

	fullPath := filepath.Join(rootFolder, string(pathData))
	fi, fe := os.Stat(fullPath)
	if os.IsNotExist(fe) {
		ctx.Error(gtype.ErrInternal, fe)
		return
	}

	if fi.IsDir() {
		fileName := fmt.Sprintf("%s.zip", filepath.Base(fullPath))
		ctx.Response().Header().Set("Content-Disposition", fmt.Sprint("attachment; filename=", fileName))
		s.compressFolder(ctx.Response(), fullPath, "", nil)
	} else {
		cfgFile, le := os.OpenFile(fullPath, os.O_RDONLY, 0666)
		if le != nil {
			ctx.Error(gtype.ErrInternal, le)
			return
		}
		defer cfgFile.Close()

		fileName := filepath.Base(fullPath)
		ctx.Response().Header().Set("Content-Disposition", fmt.Sprint("attachment; filename=", fileName))

		contentLength := fi.Size()
		ctx.Response().Header().Set("Content-Length", fmt.Sprint(contentLength))

		io.Copy(ctx.Response(), cfgFile)
	}
}

func (s *Service) DownloadTomcatConfigFileDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "下载应用配置文件")
	function.SetNote("返回应用配置文本内容")
	function.SetRemark("如果指定的路径为文件夹，则返回文件夹的压缩内容")
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) ModTomcatConfig(ctx gtype.Context, ps gtype.Params) {
	name := strings.TrimSpace(ctx.Request().FormValue("name"))
	if len(name) < 1 {
		ctx.Error(gtype.ErrInput, "服务名称(name)为空")
		return
	}
	path := strings.TrimSpace(ctx.Request().FormValue("path"))
	if len(path) > 0 {
		pathData, err := base64.URLEncoding.DecodeString(path)
		if err != nil {
			ctx.Error(gtype.ErrInput, fmt.Sprintf("路径(path=%s)不是有效的base字符串: ", path), err)
			return
		}
		path = string(pathData)
	}

	info := s.cfg.Sys.Svc.GetTomcatByServiceName(name)
	if info == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("服务(%s)不存在", name))
		return
	}

	rootFolder := info.WebCfg
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "配置物理根路径为空")
		return
	}
	fullPath := filepath.Join(rootFolder, path)
	fi, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("路径(%s)不存在: ", path), err)
		return
	}
	folderPath := fullPath
	if !fi.IsDir() {
		folderPath = filepath.Dir(fullPath)
	}

	uploadFile, head, err := ctx.Request().FormFile("file")
	if err != nil {
		ctx.Error(gtype.ErrInput, "上传文件无效: ", err)
		return
	}
	defer uploadFile.Close()

	buf := &bytes.Buffer{}
	fileSize, err := buf.ReadFrom(uploadFile)
	if err != nil {
		ctx.Error(gtype.ErrInput, "读取文件上传文件失败: ", err)
		return
	}
	if fileSize < 1 {
		ctx.Error(gtype.ErrInput, "上传的文件无效: 文件大小为0")
		return
	}

	targetFilePath := filepath.Join(folderPath, filepath.Base(head.Filename))
	err = os.RemoveAll(targetFilePath)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "删除源文件失败: ", err)
		return
	}

	fileWriter, err := os.Create(targetFilePath)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "创建文件失败: ", err)
		return
	}
	defer fileWriter.Close()

	uploadFile.Seek(0, 0)
	_, err = io.Copy(fileWriter, uploadFile)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "保存文件失败: ", err)
		return
	}

	argument := &gmodel.ServerArgument{
		Name: name,
	}

	go s.writeOptMessage(gtype.WSTomcatCfgUpdated, argument)

	ctx.Success(argument)
}

func (s *Service) ModTomcatConfigDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "上传应用配置")
	function.SetNote("上传应用程序配置文件")
	function.AddInputHeader(true, "content-type", "内容类型", gtype.ContentTypeFormData)
	function.AddInputForm(true, "name", "tomcat服务名称", gtype.FormValueKindText, "")
	function.AddInputForm(true, "path", "路径", gtype.FormValueKindText, "")
	function.AddInputForm(true, "file", "应用程序配置文件", gtype.FormValueKindFile, nil)
	function.SetOutputDataExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) CreateTomcatConfigFolder(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServiceTomcatCfgFolder{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInput, "服务名称(name)为空")
		return
	}
	if len(argument.Folder) < 1 {
		ctx.Error(gtype.ErrInput, "文件夹名称(folder)为空")
		return
	}
	path := ""
	if len(argument.Path) > 0 {
		pathData, pe := base64.URLEncoding.DecodeString(argument.Path)
		if pe != nil {
			ctx.Error(gtype.ErrInput, fmt.Sprintf("路径(%s)不是有效的base字符串: ", argument.Path), pe)
			return
		}
		path = string(pathData)
	}

	info := s.cfg.Sys.Svc.GetTomcatByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}
	rootFolder := info.WebCfg
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "配置物理根路径为空")
		return
	}

	fullPath := filepath.Join(rootFolder, path, argument.Folder)
	err = os.MkdirAll(fullPath, 0777)
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("创建文件夹'%s'失败: ", argument.Folder), err)
		return
	}

	args := &gmodel.ServerArgument{
		Name: argument.Name,
	}

	go s.writeOptMessage(gtype.WSTomcatCfgUpdated, args)

	ctx.Success(args)
}

func (s *Service) CreateTomcatConfigFolderDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "新建应用配置文件夹")
	function.SetInputJsonExample(&gmodel.ServiceTomcatCfgFolder{
		Name: "example",
	})
	function.SetOutputDataExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) DeleteTomcatConfig(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServiceTomcatCfgFolder{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInput, "服务名称(name)为空")
		return
	}
	if len(argument.Path) < 1 {
		ctx.Error(gtype.ErrInput, "路径(path)为空")
		return
	}
	pathData, pe := base64.URLEncoding.DecodeString(argument.Path)
	if pe != nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("路径(%s)不是有效的base64字符串: ", argument.Path), pe)
		return
	}
	path := string(pathData)
	if len(path) < 1 {
		ctx.Error(gtype.ErrInput, "路径为空")
		return
	}

	info := s.cfg.Sys.Svc.GetTomcatByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}
	rootFolder := info.WebCfg
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "配置物理根路径为空")
		return
	}

	fullPath := filepath.Join(rootFolder, path)
	err = os.RemoveAll(fullPath)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "删除失败: ", err)
		return
	}

	args := &gmodel.ServerArgument{
		Name: argument.Name,
	}

	go s.writeOptMessage(gtype.WSTomcatCfgDeleted, args)

	ctx.Success(args)
}

func (s *Service) DeleteTomcatConfigDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "删除应用配置")
	function.SetInputJsonExample(&gmodel.ServiceTomcatCfgFolder{
		Name: "example",
	})
	function.SetOutputDataExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) GetTomcatLogs(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServerArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInternal, "服务名称(name)为空")
		return
	}

	info := s.cfg.Sys.Svc.GetTomcatByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}

	cfg := &gmodel.ServiceLogFile{}
	s.getTomcatLog(cfg, info.WebLog, "")

	cfg.Sort()
	ctx.Success(cfg.Children)
}

func (s *Service) GetTomcatLogsDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "获取服务日志列表")
	function.SetInputJsonExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.SetOutputDataExample([]*gmodel.ServiceLogFile{
		{
			Name:     "example",
			Path:     base64.URLEncoding.EncodeToString([]byte("example")),
			Children: []*gmodel.ServiceLogFile{},
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) ViewTomcatLogFile(ctx gtype.Context, ps gtype.Params) {
	name := ps.ByName("name")
	if len(name) < 1 {
		ctx.Error(gtype.ErrInput, "服务名称(name)为空")
		return
	}

	pathName := ps.ByName("path")
	if len(pathName) < 1 {
		ctx.Error(gtype.ErrInput, "base64路径为空")
		return
	}

	pathData, err := base64.URLEncoding.DecodeString(pathName)
	if err != nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("路径base64(%s)无效", pathName))
		return
	}
	if len(pathData) < 1 {
		ctx.Error(gtype.ErrInput, "路径为空")
		return
	}

	info := s.cfg.Sys.Svc.GetTomcatByServiceName(name)
	if info == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("服务(%s)不存在", name))
		return
	}

	rootFolder := info.WebLog
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "日志物理根路径为空")
		return
	}

	fullPath := filepath.Join(rootFolder, string(pathData))
	fi, fe := os.Stat(fullPath)
	if os.IsNotExist(fe) {
		ctx.Error(gtype.ErrInternal, fe)
		return
	}
	if fi.IsDir() {
		ctx.Error(gtype.ErrInternal, "指定的路径为文件夹")
		return
	}

	cfgFile, le := os.OpenFile(fullPath, os.O_RDONLY, 0666)
	if le != nil {
		ctx.Error(gtype.ErrInternal, le)
		return
	}
	defer cfgFile.Close()

	extName := strings.ToLower(path.Ext(fullPath))
	if extName == ".xml" {
		ctx.Response().Header().Set("Content-Type", "application/xml;charset=utf-8")
	} else if extName == ".json" {
		ctx.Response().Header().Set("Content-Type", "application/json;charset=utf-8")
	} else {
		ctx.Response().Header().Set("Content-Type", "text/plain;charset=utf-8")
	}

	contentLength := fi.Size()
	ctx.Response().Header().Set("Content-Length", fmt.Sprint(contentLength))

	io.Copy(ctx.Response(), cfgFile)
}

func (s *Service) ViewTomcatLogFileDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "查看服务日志文件")
	function.SetNote("返回服务日志文本内容")
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) DownloadTomcatLogFile(ctx gtype.Context, ps gtype.Params) {
	name := ps.ByName("name")
	if len(name) < 1 {
		ctx.Error(gtype.ErrInput, "服务名称(name)为空")
		return
	}

	pathName := ps.ByName("path")
	if len(pathName) < 1 {
		ctx.Error(gtype.ErrInput, "base64路径为空")
		return
	}

	pathData, err := base64.URLEncoding.DecodeString(pathName)
	if err != nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("路径base64(%s)无效", pathName))
		return
	}
	if len(pathData) < 1 {
		ctx.Error(gtype.ErrInput, "路径为空")
		return
	}

	info := s.cfg.Sys.Svc.GetTomcatByServiceName(name)
	if info == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("服务(%s)不存在", name))
		return
	}

	rootFolder := info.WebLog
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "日志物理根路径为空")
		return
	}

	fullPath := filepath.Join(rootFolder, string(pathData))
	fi, fe := os.Stat(fullPath)
	if os.IsNotExist(fe) {
		ctx.Error(gtype.ErrInternal, fe)
		return
	}

	if fi.IsDir() {
		fileName := fmt.Sprintf("%s.zip", filepath.Base(fullPath))
		ctx.Response().Header().Set("Content-Disposition", fmt.Sprint("attachment; filename=", fileName))
		s.compressFolder(ctx.Response(), fullPath, "", nil)
	} else {
		cfgFile, le := os.OpenFile(fullPath, os.O_RDONLY, 0666)
		if le != nil {
			ctx.Error(gtype.ErrInternal, le)
			return
		}
		defer cfgFile.Close()

		fileName := filepath.Base(fullPath)
		ctx.Response().Header().Set("Content-Disposition", fmt.Sprint("attachment; filename=", fileName))

		contentLength := fi.Size()
		ctx.Response().Header().Set("Content-Length", fmt.Sprint(contentLength))

		io.Copy(ctx.Response(), cfgFile)
	}
}

func (s *Service) DownloadTomcatLogFileDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "下载服务日志文件")
	function.SetNote("返回应服务日志文本内容")
	function.SetRemark("如果指定的路径为文件夹，则返回文件夹的压缩内容")
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) DeleteTomcatLog(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServiceTomcatCfgFolder{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInput, "服务名称(name)为空")
		return
	}
	if len(argument.Path) < 1 {
		ctx.Error(gtype.ErrInput, "路径(path)为空")
		return
	}
	pathData, pe := base64.URLEncoding.DecodeString(argument.Path)
	if pe != nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("路径(%s)不是有效的base64字符串: ", argument.Path), pe)
		return
	}
	path := string(pathData)
	if len(path) < 1 {
		ctx.Error(gtype.ErrInput, "路径为空")
		return
	}

	info := s.cfg.Sys.Svc.GetTomcatByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}
	rootFolder := info.WebLog
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "日志物理根路径为空")
		return
	}

	fullPath := filepath.Join(rootFolder, path)
	err = os.RemoveAll(fullPath)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "删除失败: ", err)
		return
	}

	args := &gmodel.ServerArgument{
		Name: argument.Name,
	}

	ctx.Success(args)
}

func (s *Service) DeleteTomcatLogDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogTomcat)
	function := catalog.AddFunction(method, uri, "删除服务日志")
	function.SetInputJsonExample(&gmodel.ServiceTomcatCfgFolder{
		Name: "example",
	})
	function.SetOutputDataExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) getTomcatCfg(cfg *gmodel.ServiceTomcatCfg, root, folder string) {
	if cfg == nil {
		return
	}
	if cfg.Children == nil {
		cfg.Children = make([]*gmodel.ServiceTomcatCfg, 0)
	}

	if len(root) < 1 {
		return
	}

	fs, fe := ioutil.ReadDir(filepath.Join(root, folder))
	if fe != nil {
		return
	}

	for _, f := range fs {
		name := f.Name()
		path := filepath.Join(folder, name)
		item := &gmodel.ServiceTomcatCfg{
			Name:     name,
			Path:     base64.URLEncoding.EncodeToString([]byte(path)),
			Children: []*gmodel.ServiceTomcatCfg{},
		}
		cfg.Children = append(cfg.Children, item)
		if f.IsDir() {
			item.Folder = true
			s.getTomcatCfg(item, root, path)
		}
	}
}

func (s *Service) getTomcatLog(cfg *gmodel.ServiceLogFile, root, folder string) {
	if cfg == nil {
		return
	}
	if cfg.Children == nil {
		cfg.Children = make([]*gmodel.ServiceLogFile, 0)
	}

	if len(root) < 1 {
		return
	}

	fs, fe := ioutil.ReadDir(filepath.Join(root, folder))
	if fe != nil {
		return
	}

	for _, f := range fs {
		name := f.Name()
		path := filepath.Join(folder, name)
		item := &gmodel.ServiceLogFile{
			Name:     name,
			Path:     base64.URLEncoding.EncodeToString([]byte(path)),
			ModTime:  gtype.DateTime(f.ModTime()),
			Children: []*gmodel.ServiceLogFile{},
		}
		cfg.Children = append(cfg.Children, item)
		if f.IsDir() {
			item.Folder = true
			s.getTomcatLog(item, root, path)
		} else {
			item.Size = f.Size()
			item.SizeText = s.sizeToText(float64(item.Size))
		}
	}
}

func (s *Service) getTomcatAppVersion(folder string) string {
	path, err := s.getFilePath(folder, "pom.xml")
	if err != nil {
		return ""
	}
	if len(path) < 1 {
		return ""
	}

	pom := &gmodel.ServiceTomcatPom{}
	err = pom.LoadFromFile(path)
	if err != nil {
		return ""
	}
	if len(pom.Parent.Version) > 0 {
		return pom.Parent.Version
	}

	return pom.ModelVersion
}
