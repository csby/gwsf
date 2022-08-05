package controller

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/csby/gwsf/gfile"
	"github.com/csby/gwsf/gmodel"
	"github.com/csby/gwsf/gtype"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func (s *Service) StartNginx(ctx gtype.Context, ps gtype.Params) {
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
	info := s.cfg.Sys.Svc.GetNginxByServiceName(argument.Name)
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

func (s *Service) StartNginxDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogNginx)
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

func (s *Service) StopNginx(ctx gtype.Context, ps gtype.Params) {
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
	info := s.cfg.Sys.Svc.GetNginxByServiceName(argument.Name)
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

func (s *Service) StopNginxDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogNginx)
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

func (s *Service) RestartNginx(ctx gtype.Context, ps gtype.Params) {
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
	info := s.cfg.Sys.Svc.GetNginxByServiceName(argument.Name)
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

func (s *Service) RestartNginxDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogNginx)
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

func (s *Service) GetNginxes(ctx gtype.Context, ps gtype.Params) {
	results := make([]*gmodel.ServiceNginxInfo, 0)

	if s.cfg != nil {
		items := s.cfg.Sys.Svc.Nginxes
		c := len(items)
		for i := 0; i < c; i++ {
			item := items[i]
			if item == nil {
				continue
			}
			if len(item.ServiceName) < 1 {
				continue
			}

			result := &gmodel.ServiceNginxInfo{
				Name:        item.Name,
				ServiceName: item.ServiceName,
				Remark:      item.Remark,
				Locations:   make([]*gmodel.ServiceNginxLocation, 0),
			}
			if len(result.Name) < 1 {
				result.Name = result.ServiceName
			}
			result.Status, _ = s.getStatus(result.ServiceName)

			lc := len(item.Locations)
			for li := 0; li < lc; li++ {
				l := item.Locations[li]
				if l == nil {
					continue
				}

				location := &gmodel.ServiceNginxLocation{
					Name: l.Name,
					Root: l.Root,
					Urls: make([]string, 0),
				}
				location.Version, location.DeployTime, _ = s.getNginxAppInfo(l.Root)

				uc := len(l.Urls)
				for ui := 0; ui < uc; ui++ {
					location.Urls = append(location.Urls, l.Urls[ui])
				}

				result.Locations = append(result.Locations, location)
			}

			results = append(results, result)
		}
	}

	ctx.Success(results)
}

func (s *Service) GetNginxesDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogNginx)
	function := catalog.AddFunction(method, uri, "获取服务列表")
	function.SetOutputDataExample([]*gmodel.ServiceNginxInfo{
		{
			Name:        "example",
			ServiceName: "nginx",
			Locations: []*gmodel.ServiceNginxLocation{
				{
					Name: "api",
					Urls: []string{},
				},
			},
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) ModNginxApp(ctx gtype.Context, ps gtype.Params) {
	svcName := strings.TrimSpace(ctx.Request().FormValue("svcName"))
	if len(svcName) < 1 {
		ctx.Error(gtype.ErrInput, "服务名称(svcName)为空")
		return
	}
	appName := strings.TrimSpace(ctx.Request().FormValue("appName"))
	if len(appName) < 1 {
		ctx.Error(gtype.ErrInput, "站点名称(appName)为空")
		return
	}

	info := s.cfg.Sys.Svc.GetNginxByServiceName(svcName)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务名称(%s)不存在", svcName))
		return
	}

	app := info.GetLocationByName(appName)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("站点名称(%s)不存在", appName))
		return
	}

	rootFolder := app.Root
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "站点服务物理根路径为空")
		return
	}

	uploadFile, _, err := ctx.Request().FormFile("file")
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

	appFolder := rootFolder
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

	argument := &gmodel.ServiceNginxArgument{
		SvcName: svcName,
		AppName: appName,
	}
	argument.Version, argument.DeployTime, _ = s.getNginxAppInfo(rootFolder)

	go s.writeOptMessage(gtype.WSNginxAppUpdated, argument)

	ctx.Success(argument)
}

func (s *Service) ModNginxAppDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogNginx)
	function := catalog.AddFunction(method, uri, "上传站点程序")
	function.SetNote("上传应用程序打包文件(.war, .zip或.tar.gz)，并解压缩到站点根目录下，成功时返回服务名称及站点名称信息")
	function.SetRemark("压缩包内的文件应为网站内容文件，不要嵌套在文件夹中")
	function.AddInputHeader(true, "content-type", "内容类型", gtype.ContentTypeFormData)
	function.AddInputForm(true, "svcName", "服务名称", gtype.FormValueKindText, "")
	function.AddInputForm(true, "appName", "站点名称", gtype.FormValueKindText, "")
	function.AddInputForm(true, "file", "应用程序打包文件(.war, .zip或.tar.gz)", gtype.FormValueKindFile, nil)
	function.SetOutputDataExample(&gmodel.ServiceNginxArgument{
		SvcName: "nginx",
		AppName: "api",
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) GetNginxAppDetail(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServiceNginxDetailArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.SvcName) < 1 {
		ctx.Error(gtype.ErrInternal, "服务名称(svcName)为空")
		return
	}
	if len(argument.AppName) < 1 {
		ctx.Error(gtype.ErrInternal, "站点名称(appName)为空")
		return
	}

	info := s.cfg.Sys.Svc.GetNginxByServiceName(argument.SvcName)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务名称(%s)不存在", argument.SvcName))
		return
	}

	app := info.GetLocationByName(argument.AppName)
	if app == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("站点名称(%s)不存在", argument.AppName))
		return
	}

	cfg := &gmodel.FileInfo{}
	s.getFileInfos(cfg, app.Root)

	cfg.Sort()
	ctx.Success(cfg)
}

func (s *Service) GetNginxAppDetailDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogNginx)
	function := catalog.AddFunction(method, uri, "获取站点程序详细信息")
	function.SetInputJsonExample(&gmodel.ServiceNginxDetailArgument{
		SvcName: "nginx",
		AppName: "api",
	})
	function.SetOutputDataExample(&gmodel.FileInfo{})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) GetNginxLogs(ctx gtype.Context, ps gtype.Params) {
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

	info := s.cfg.Sys.Svc.GetNginxByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}

	cfg := &gmodel.ServiceLogFile{}
	s.getLogFiles(cfg, info.Log, "")

	cfg.Sort()
	ctx.Success(cfg.Children)
}

func (s *Service) GetNginxLogsDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogNginx)
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

func (s *Service) ViewNginxLogFile(ctx gtype.Context, ps gtype.Params) {
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

	info := s.cfg.Sys.Svc.GetNginxByServiceName(name)
	if info == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("服务(%s)不存在", name))
		return
	}

	rootFolder := info.Log
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

func (s *Service) ViewNginxLogFileDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogNginx)
	function := catalog.AddFunction(method, uri, "查看服务日志文件")
	function.SetNote("返回服务日志文本内容")
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) DownloadNginxLogFile(ctx gtype.Context, ps gtype.Params) {
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

	info := s.cfg.Sys.Svc.GetNginxByServiceName(name)
	if info == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("服务(%s)不存在", name))
		return
	}

	rootFolder := info.Log
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

func (s *Service) DownloadNginxLogFileDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogNginx)
	function := catalog.AddFunction(method, uri, "下载服务日志文件")
	function.SetNote("返回应服务日志文本内容")
	function.SetRemark("如果指定的路径为文件夹，则返回文件夹的压缩内容")
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) DeleteNginxLog(ctx gtype.Context, ps gtype.Params) {
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

func (s *Service) DeleteNginxLogDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogNginx)
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

func (s Service) getNginxAppInfo(folder string) (version, deployTime string, err error) {
	version = ""
	deployTime = ""
	err = nil
	if len(folder) < 1 {
		return
	}
	fi, fe := os.Stat(folder)
	if fe != nil {
		err = fe
		return
	}
	if !fi.IsDir() {
		return
	}
	deployTime = fi.ModTime().Format("2006-01-02 15:04:05")

	version, _ = s.getTextVersion(folder)
	if len(version) < 1 {
		version, _ = s.getJsonVersion(folder)
		if len(version) < 1 {
			version = s.getNginxAppVersion(folder)
		}
	}

	return
}

func (s *Service) getNginxAppVersion(folder string) string {
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
