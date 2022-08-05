package controller

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/csby/gwsf/gfile"
	"github.com/csby/gwsf/gmodel"
	"github.com/csby/gwsf/gtype"
	"github.com/kardianos/service"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

func (s *Service) GetCustoms(ctx gtype.Context, ps gtype.Params) {
	results := s.getCustoms()
	c := len(results)
	for i := 0; i < c; i++ {
		item := results[i]
		if item == nil {
			continue
		}

		item.Status, _ = s.getStatus(item.GetServiceName())
	}

	sort.Sort(results)
	ctx.Success(results)
}

func (s *Service) GetCustomsDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
	function := catalog.AddFunction(method, uri, "获取服务列表")
	function.SetOutputDataExample([]*gmodel.ServiceCustomInfo{
		{
			Name:        "example",
			ServiceName: "svc-cst-example",
			DisplayName: "自定义服务示例",
			DeployTime:  gtype.DateTime(time.Now()),
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) AddCustom(ctx gtype.Context, ps gtype.Params) {
	rootFolder := s.cfg.Sys.Svc.Custom.App
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
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

	infoFilePath, err := s.getCustomInfoPath(tempFolder)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	info := &gmodel.ServiceCustomInfo{}
	err = info.LoadFromFile(infoFilePath)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "读取信息文件(info.json)失败: ", err)
		return
	}
	if len(info.Name) < 1 {
		ctx.Error(gtype.ErrInput, "信息文件(info.json)中的名称(name)为空")
		return
	}

	srvFolder := filepath.Join(rootFolder, info.Name)
	_, err = os.Stat(srvFolder)
	if !os.IsNotExist(err) {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("服务(%s)已存在", info.Name))
		return
	}

	err = gfile.Copy(filepath.Dir(infoFilePath), srvFolder)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "拷贝文件夹失败: ", err)
		return
	}

	info.ServiceName = info.GetServiceName()
	info.DeployTime = gtype.DateTime(time.Now())
	info.Folder = srvFolder
	info.Status, _ = s.getStatus(info.GetServiceName())
	if len(info.DisplayName) < 1 {
		info.DisplayName = info.Name
	}
	go s.writeOptMessage(gtype.WSCustomSvcAdded, info)

	if info.Status == gtype.ServerStatusRunning {
		s.stop(info.GetServiceName())
		time.Sleep(5 * time.Second)
	}

	if info.Status == gtype.ServerStatusUnknown {
		err = s.installCustom(info)
		if err != nil {
			ctx.Error(gtype.ErrInternal, "上传成功, 但安装失败: ", err)
			return
		}
	}

	err = s.start(info.GetServiceName())
	if err != nil {
		ctx.Error(gtype.ErrInternal, "上传并安装成功, 但启动失败: ", err)
		return
	}

	svcStatus := &gmodel.ServiceStatus{Name: info.GetServiceName()}
	svcStatus.Status, err = s.getStatus(info.GetServiceName())
	if err == nil {
		go s.writeOptMessage(gtype.WSSvcStatusChanged, svcStatus)
	}

	ctx.Success(info)
}

func (s *Service) AddCustomDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
	function := catalog.AddFunction(method, uri, "添加服务")
	function.SetNote("上传服务程序打包文件(.zip或.tar.gz)，并安装成系统服务，成功时返回服务信息")
	function.SetRemark("打包文件中的根目录必须包含服务信息文件(info.json)，且服务名称(name)和可执行程序(exec)不能能为空")
	function.AddInputHeader(true, "content-type", "内容类型", gtype.ContentTypeFormData)
	function.AddInputForm(true, "file", "服务程序打包文件(.zip或.tar.gz)", gtype.FormValueKindFile, nil)
	function.SetOutputDataExample(&gmodel.ServiceCustomInfo{
		Name:        "example",
		ServiceName: "svc-cst-example",
		DisplayName: "自定义服务示例",
		DeployTime:  gtype.DateTime(time.Now()),
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) ModCustom(ctx gtype.Context, ps gtype.Params) {
	name := strings.TrimSpace(ctx.Request().FormValue("name"))
	if len(name) < 1 {
		ctx.Error(gtype.ErrInput, "服务名称(name)为空")
		return
	}

	rootFolder := s.cfg.Sys.Svc.Custom.App
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
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

	infoFilePath, err := s.getCustomInfoPath(tempFolder)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	info := &gmodel.ServiceCustomInfo{}
	err = info.LoadFromFile(infoFilePath)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "读取信息文件(info.json)失败: ", err)
		return
	}
	if len(info.Name) < 1 {
		ctx.Error(gtype.ErrInput, "信息文件(info.json)中的名称(name)为空")
		return
	}
	if info.Name != name {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("信息文件(info.json)中的名称(%s)和目标服务名称(%s)不一致", info.Name, name))
		return
	}

	svcStatus := &gmodel.ServiceStatus{Name: info.GetServiceName()}
	svcStatus.Status, err = s.getStatus(info.GetServiceName())
	if err == nil {
		if svcStatus.Status == gtype.ServerStatusRunning {
			err = s.stop(info.GetServiceName())
			if err != nil {
				ctx.Error(gtype.ErrInternal, "停止服务失败: ", err)
				return
			}

			time.Sleep(500 * time.Millisecond)
			svcStatus.Status, err = s.getStatus(info.GetServiceName())
			if err == nil {
				go s.writeOptMessage(gtype.WSSvcStatusChanged, svcStatus)
			}
			time.Sleep(3 * time.Second)
		}
	}

	srvFolder := filepath.Join(rootFolder, info.Name)
	err = os.RemoveAll(srvFolder)
	if err != nil {
		time.Sleep(3 * time.Second)
		err = os.RemoveAll(srvFolder)
		if err != nil {
			time.Sleep(5 * time.Second)
			if err != nil {
				err = os.RemoveAll(srvFolder)
				ctx.Error(gtype.ErrInternal, "删除原服务文件夹失败: ", err)
				return
			}
		}
	}

	err = gfile.Copy(filepath.Dir(infoFilePath), srvFolder)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "拷贝文件夹失败: ", err)
		return
	}

	info.ServiceName = info.GetServiceName()
	info.DeployTime = gtype.DateTime(time.Now())
	info.Folder = srvFolder
	if len(info.DisplayName) < 1 {
		info.DisplayName = info.Name
	}
	go s.writeOptMessage(gtype.WSCustomSvcUpdated, info)

	if svcStatus.Status == gtype.ServerStatusStopped {
		err = s.start(info.GetServiceName())
		if err != nil {
			ctx.Error(gtype.ErrInternal, "更新成功, 但启动失败: ", err)
			return
		}
	}
	svcStatus.Status, err = s.getStatus(info.GetServiceName())
	if err == nil {
		go s.writeOptMessage(gtype.WSSvcStatusChanged, svcStatus)
	}

	ctx.Success(info)
}

func (s *Service) ModCustomDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
	function := catalog.AddFunction(method, uri, "更新服务")
	function.SetNote("上传服务程序打包文件(.zip或.tar.gz)，并安装成系统服务，成功时返回服务信息")
	function.SetRemark("打包文件中的根目录必须包含服务信息文件(info.json)，且服务名称(name)和可执行程序(exec)不能能为空")
	function.AddInputHeader(true, "content-type", "内容类型", gtype.ContentTypeFormData)
	function.AddInputForm(true, "name", "目标服务名称(需和info.json中的name一致)", gtype.FormValueKindText, "")
	function.AddInputForm(true, "file", "服务程序打包文件(.zip或.tar.gz)", gtype.FormValueKindFile, nil)
	function.SetOutputDataExample(&gmodel.ServiceCustomInfo{
		Name:        "example",
		ServiceName: "svc-cst-example",
		DisplayName: "自定义服务示例",
		DeployTime:  gtype.DateTime(time.Now()),
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) DelCustom(ctx gtype.Context, ps gtype.Params) {
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

	rootFolder := s.cfg.Sys.Svc.Custom.App
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
		return
	}

	infoFolder := filepath.Join(rootFolder, argument.Name)
	infoPath := filepath.Join(infoFolder, "info.json")
	info := &gmodel.ServiceCustomInfo{}
	err = info.LoadFromFile(infoPath)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "读取信息文件(info.json)失败: ", err)
		return
	}
	if len(info.Name) < 1 {
		ctx.Error(gtype.ErrInput, "信息文件(info.json)中的名称(name)为空")
		return
	}
	info.Folder = infoFolder

	svcStatus := &gmodel.ServiceStatus{Name: info.GetServiceName()}
	svcStatus.Status, err = s.getStatus(info.GetServiceName())
	if err != nil {
		ctx.Error(gtype.ErrInternal, "获取服务状态失败: ", err)
		return
	}
	if svcStatus.Status != gtype.ServerStatusUnknown {
		ctx.Error(gtype.ErrInternal, "服务未卸载")
		return
	}

	err = os.RemoveAll(infoFolder)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "删除原服务文件夹失败: ", err)
		return
	}

	go s.writeOptMessage(gtype.WSCustomSvcDeleted, argument)

	ctx.Success(argument)
}

func (s *Service) DelCustomDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
	function := catalog.AddFunction(method, uri, "删除服务")
	function.SetRemark("已卸载服务才能删除")
	function.SetInputJsonExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) DownloadCustom(ctx gtype.Context, ps gtype.Params) {
	name := ps.ByName("name")
	if len(name) < 1 {
		ctx.Error(gtype.ErrInput, "名称为空")
		return
	}

	rootFolder := s.cfg.Sys.Svc.Custom.App
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
		return
	}

	svcFolder := filepath.Join(rootFolder, name)
	infoFilePath := filepath.Join(svcFolder, "info.json")
	info := &gmodel.ServiceCustomInfo{}
	err := info.LoadFromFile(infoFilePath)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "读取信息文件(info.json)失败: ", err)
		return
	}
	if len(info.Name) < 1 {
		ctx.Error(gtype.ErrInput, "信息文件(info.json)中的名称(name)为空")
		return
	}

	fileName := fmt.Sprintf("%s.zip", info.Name)
	ctx.Response().Header().Set("Content-Disposition", fmt.Sprint("attachment; filename=", fileName))

	s.compressFolder(ctx.Response(), svcFolder, "", nil)
}

func (s *Service) DownloadCustomDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
	function := catalog.AddFunction(method, uri, "下载服务")
	function.SetNote("服务程序文件(.zip)")
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) GetCustomDetail(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServerArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInternal, "名称(name)为空")
		return
	}

	rootFolder := s.cfg.Sys.Svc.Custom.App
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
		return
	}

	svcFolder := filepath.Join(rootFolder, argument.Name)
	cfg := &gmodel.FileInfo{}
	s.getFileInfos(cfg, svcFolder)

	cfg.Sort()
	ctx.Success(cfg)
}

func (s *Service) GetCustomDetailDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
	function := catalog.AddFunction(method, uri, "获取服务详细信息")
	function.SetInputJsonExample(&gmodel.ServerArgument{
		Name: "svc",
	})
	function.SetOutputDataExample(&gmodel.FileInfo{})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) InstallCustom(ctx gtype.Context, ps gtype.Params) {
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

	rootFolder := s.cfg.Sys.Svc.Custom.App
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
		return
	}

	infoFolder := filepath.Join(rootFolder, argument.Name)
	info, ge := s.getCustomInfo(infoFolder)
	if ge != nil {
		ctx.Error(ge)
		return
	}

	err = s.installCustom(info)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	svcStatus := &gmodel.ServiceStatus{Name: info.GetServiceName()}
	svcStatus.Status, err = s.getStatus(info.GetServiceName())
	if err == nil {
		go s.writeOptMessage(gtype.WSSvcStatusChanged, svcStatus)
	}

	ctx.Success(info)
}

func (s *Service) InstallCustomDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
	function := catalog.AddFunction(method, uri, "安装服务")
	function.SetInputJsonExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) UninstallCustom(ctx gtype.Context, ps gtype.Params) {
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

	rootFolder := s.cfg.Sys.Svc.Custom.App
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
		return
	}

	infoFolder := filepath.Join(rootFolder, argument.Name)
	info, ge := s.getCustomInfo(infoFolder)
	if ge != nil {
		ctx.Error(ge)
		return
	}

	err = s.uninstall(info.GetServiceName())
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}
	svcStatus := &gmodel.ServiceStatus{Name: info.GetServiceName()}
	svcStatus.Status, err = s.getStatus(info.GetServiceName())
	go s.writeOptMessage(gtype.WSSvcStatusChanged, svcStatus)

	ctx.Success(info)
}

func (s *Service) UninstallCustomDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
	function := catalog.AddFunction(method, uri, "卸载服务")
	function.SetInputJsonExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) StartCustom(ctx gtype.Context, ps gtype.Params) {
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

	rootFolder := s.cfg.Sys.Svc.Custom.App
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
		return
	}

	infoFolder := filepath.Join(rootFolder, argument.Name)
	info, ge := s.getCustomInfo(infoFolder)
	if ge != nil {
		ctx.Error(ge)
		return
	}

	err = s.start(info.GetServiceName())
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	svcStatus := &gmodel.ServiceStatus{Name: info.GetServiceName()}
	svcStatus.Status, err = s.getStatus(info.GetServiceName())
	if err == nil {
		go s.writeOptMessage(gtype.WSSvcStatusChanged, svcStatus)
	}

	ctx.Success(info)
}

func (s *Service) StartCustomDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
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

func (s *Service) StopCustom(ctx gtype.Context, ps gtype.Params) {
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

	rootFolder := s.cfg.Sys.Svc.Custom.App
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
		return
	}

	infoFolder := filepath.Join(rootFolder, argument.Name)
	info, ge := s.getCustomInfo(infoFolder)
	if ge != nil {
		ctx.Error(ge)
		return
	}

	err = s.stop(info.GetServiceName())
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}
	svcStatus := &gmodel.ServiceStatus{Name: info.GetServiceName()}
	svcStatus.Status, err = s.getStatus(info.GetServiceName())
	if err == nil {
		go s.writeOptMessage(gtype.WSSvcStatusChanged, svcStatus)
	}

	ctx.Success(info)
}

func (s *Service) StopCustomDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
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

func (s *Service) RestartCustom(ctx gtype.Context, ps gtype.Params) {
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

	rootFolder := s.cfg.Sys.Svc.Custom.App
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
		return
	}

	infoFolder := filepath.Join(rootFolder, argument.Name)
	info, ge := s.getCustomInfo(infoFolder)
	if ge != nil {
		ctx.Error(ge)
		return
	}

	err = s.restart(info.GetServiceName())
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}
	svcStatus := &gmodel.ServiceStatus{Name: info.GetServiceName()}
	svcStatus.Status, err = s.getStatus(info.GetServiceName())
	if err == nil {
		go s.writeOptMessage(gtype.WSSvcStatusChanged, svcStatus)
	}

	ctx.Success(info)
}

func (s *Service) RestartCustomDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
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

func (s *Service) GetCustomLogFiles(ctx gtype.Context, ps gtype.Params) {
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

	rootFolder := s.cfg.Sys.Svc.Custom.Log
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
		return
	}
	results := s.getFiles(rootFolder, argument.Name)

	appLogs := s.getFiles(filepath.Join(s.cfg.Sys.Svc.Custom.App, argument.Name), "log")
	count := len(appLogs)
	for index := 0; index < count; index++ {
		item := appLogs[index]
		path, pe := base64.URLEncoding.DecodeString(item.Path)
		if pe == nil {
			item.Path = base64.URLEncoding.EncodeToString([]byte(filepath.Join(argument.Name, string(path))))
			results = append(results, item)
		}
	}

	sort.Sort(results)
	ctx.Success(results)
}

func (s *Service) GetCustomLogFilesDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
	function := catalog.AddFunction(method, uri, "获取服务日志文件列表")
	function.SetInputJsonExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.SetOutputDataExample([]*gmodel.ServiceLogFile{
		{
			Name:     "2021-12-29.log",
			Size:     12783,
			SizeText: s.sizeToText(float64(12783)),
			ModTime:  gtype.DateTime(time.Now()),
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) DownloadCustomLogFile(ctx gtype.Context, ps gtype.Params) {
	path := ps.ByName("path")
	if len(path) < 1 {
		ctx.Error(gtype.ErrInput, "路径base64为空")
		return
	}

	pathData, err := base64.URLEncoding.DecodeString(path)
	if err != nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("路径base64(%s)无效", path))
		return
	}
	if len(pathData) < 1 {
		ctx.Error(gtype.ErrInput, "路径为空")
		return
	}

	rootFolder := s.cfg.Sys.Svc.Custom.Log
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
		return
	}

	logPath := filepath.Join(rootFolder, string(pathData))
	fi, fe := os.Stat(logPath)
	if os.IsNotExist(fe) {
		logPath2 := filepath.Join(s.cfg.Sys.Svc.Custom.App, string(pathData))
		fi2, fe2 := os.Stat(logPath2)
		if os.IsNotExist(fe2) {
			ctx.Error(gtype.ErrInternal, fe, " and ", fe2)
			return
		}
		logPath = logPath2
		fi = fi2
	}
	logFile, le := os.OpenFile(logPath, os.O_RDONLY, 0666)
	if le != nil {
		ctx.Error(gtype.ErrInternal, le)
		return
	}
	defer logFile.Close()

	contentLength := fi.Size()
	fileName := fmt.Sprintf("%s", filepath.Base(logPath))
	ctx.Response().Header().Set("Content-Disposition", fmt.Sprint("attachment; filename=", fileName))
	ctx.Response().Header().Set("Content-Length", fmt.Sprint(contentLength))

	io.Copy(ctx.Response(), logFile)
}

func (s *Service) DownloadCustomLogFileDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
	function := catalog.AddFunction(method, uri, "下载服务日志文件")
	function.SetNote("服务日志文件(.log)")
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) ViewCustomLogFile(ctx gtype.Context, ps gtype.Params) {
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

	rootFolder := s.cfg.Sys.Svc.Custom.Log
	if len(rootFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "服务物理根路径为空")
		return
	}

	logPath := filepath.Join(rootFolder, string(pathData))
	fi, fe := os.Stat(logPath)
	if os.IsNotExist(fe) {
		logPath2 := filepath.Join(s.cfg.Sys.Svc.Custom.App, string(pathData))
		fi2, fe2 := os.Stat(logPath2)
		if os.IsNotExist(fe2) {
			ctx.Error(gtype.ErrInternal, fe, " and ", fe2)
			return
		}
		logPath = logPath2
		fi = fi2
	}
	logFile, le := os.OpenFile(logPath, os.O_RDONLY, 0666)
	if le != nil {
		ctx.Error(gtype.ErrInternal, le)
		return
	}
	defer logFile.Close()

	extName := strings.ToLower(path.Ext(logPath))
	if extName == ".xml" {
		ctx.Response().Header().Set("Content-Type", "application/xml;charset=utf-8")
	} else if extName == ".json" {
		ctx.Response().Header().Set("Content-Type", "application/json;charset=utf-8")
	} else {
		ctx.Response().Header().Set("Content-Type", "text/plain;charset=utf-8")
	}

	contentLength := fi.Size()
	ctx.Response().Header().Set("Content-Length", fmt.Sprint(contentLength))

	io.Copy(ctx.Response(), logFile)
}

func (s *Service) ViewCustomLogFileDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
	function := catalog.AddFunction(method, uri, "查看服务日志文件")
	function.SetNote("返回日志文件文本内容")
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) getCustoms() gmodel.ServiceCustomInfoCollection {
	results := make(gmodel.ServiceCustomInfoCollection, 0)
	rootFolder := s.cfg.Sys.Svc.Custom.App
	if len(rootFolder) > 0 {
		fs, fe := ioutil.ReadDir(rootFolder)
		if fe == nil {
			for _, f := range fs {
				if f.IsDir() {
					infoPath := filepath.Join(rootFolder, f.Name(), "info.json")
					info := &gmodel.ServiceCustomInfo{}
					err := info.LoadFromFile(infoPath)
					if err == nil {
						info.ServiceName = info.GetServiceName()
						info.Folder = filepath.Dir(infoPath)
						info.DeployTime = gtype.DateTime(f.ModTime())
						if len(info.DisplayName) < 1 {
							info.DisplayName = info.Name
						}
						results = append(results, info)
					}
				}
			}
		}
	}

	return results
}

func (s *Service) installCustom(info *gmodel.ServiceCustomInfo) error {
	if info == nil {
		return fmt.Errorf("info is nil")
	}

	executable := filepath.Join(filepath.Dir(s.cfg.Module.Path), "gshell")
	if runtime.GOOS == "windows" {
		executable += ".exe"
	}
	logFolder := ""
	if len(s.cfg.Sys.Svc.Custom.Log) > 0 {
		logFolder = filepath.Join(s.cfg.Sys.Svc.Custom.Log, info.Name)
	}
	cfg := &service.Config{
		Name:        info.GetServiceName(),
		Description: info.Description,
		Arguments:   []string{info.Folder, logFolder},
		Executable:  executable,
	}
	cfg.DisplayName = cfg.Name

	svc, err := service.New(nil, cfg)
	if err != nil {
		return err
	}

	return svc.Install()
}

func (s *Service) getCustomInfo(folder string) (*gmodel.ServiceCustomInfo, gtype.Error) {
	path := filepath.Join(folder, "info.json")
	info := &gmodel.ServiceCustomInfo{}
	err := info.LoadFromFile(path)
	if err != nil {
		return nil, gtype.ErrInternal.SetDetail("读取信息文件(info.json)失败: ", err)
	}
	if len(info.Name) < 1 {
		return nil, gtype.ErrInput.SetDetail("信息文件(info.json)中的名称(name)为空")
	}
	info.Folder = folder

	return info, nil
}

func (s *Service) getCustomInfoPath(folderPath string) (string, error) {
	filePath := filepath.Join(folderPath, "info.json")
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		fs, e := ioutil.ReadDir(folderPath)
		if e != nil {
			return "", e
		}
		for _, f := range fs {
			if f.IsDir() {
				path, pe := s.getCustomInfoPath(filepath.Join(folderPath, f.Name()))
				if pe == nil {
					return path, nil
				}
			}
		}
		return "", fmt.Errorf("未包含服务信息文件(info.json)")
	} else {
		return filePath, nil
	}
}

func (s *Service) removeCustomLogs() {
	if s.cfg == nil {
		return
	}
	folder := s.cfg.Sys.Svc.Custom.Log
	if len(folder) < 1 {
		return
	}
	days := s.cfg.Sys.Svc.Custom.LogRetainDays
	if days < 1 {
		return
	}

	go s.doRemoveCustomLogs(folder, days)
}

func (s *Service) doRemoveCustomLogs(folder string, days int64) {
	defer func() {
		if err := recover(); err != nil {
			s.LogError("remove custom logs fail: ", err)
		}
	}()

	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	zero := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
		0, 0, 0, 0,
		tomorrow.Location())
	duration := zero.Sub(now)
	expiration := 24 * time.Hour * time.Duration(days)

	for {
		if duration > 0 {
			time.Sleep(duration)
		}
		duration = 24 * time.Hour

		fs, fe := ioutil.ReadDir(folder)
		if fe != nil {
			continue
		}

		now = time.Now()
		for _, f := range fs {
			if !f.IsDir() {
				continue
			}

			s.removeFiles(filepath.Join(folder, f.Name()), now, expiration)
		}
	}
}
