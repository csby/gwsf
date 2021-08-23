package controller

import (
	"bytes"
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gfile"
	"github.com/csby/gwsf/gtype"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type Service struct {
	controller

	svcMgr   gtype.SvcUpdMgr
	bootTime time.Time
}

func NewService(log gtype.Log, cfg *gcfg.Config, svcMgr gtype.SvcUpdMgr) *Service {
	instance := &Service{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.svcMgr = svcMgr

	if cfg != nil {
		instance.bootTime = cfg.Svc.BootTime
	} else {
		instance.bootTime = time.Now()
	}

	return instance
}

func (s *Service) Version(ctx gtype.Context, ps gtype.Params) {
	data := ""
	cfg := s.cfg
	if cfg != nil {
		data = cfg.Module.Version
	}

	ctx.Success(data)
}

func (s *Service) VersionDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "后台服务")
	function := catalog.AddFunction(method, uri, "获取服务版本号")
	function.SetNote("获取当前服务版本号信息")
	function.SetOutputDataExample("1.0.1.0")
	function.AddOutputError(gtype.ErrInternal)
}

func (s *Service) Info(ctx gtype.Context, ps gtype.Params) {
	data := &gtype.SvcInfo{BootTime: gtype.DateTime(s.bootTime)}
	cfg := s.cfg
	if cfg != nil {
		data.Name = cfg.Module.Name
		data.Version = cfg.Module.Version
		data.Remark = cfg.Module.Remark
		data.DownloadTitle = cfg.Svc.DownloadTitle
		data.DownloadUrl = cfg.Svc.DownloadUrl
	}

	ctx.Success(data)
}

func (s *Service) InfoDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "后台服务")
	function := catalog.AddFunction(method, uri, "获取服务信息")
	function.SetNote("获取当前服务信息")
	function.SetOutputDataExample(&gtype.SvcInfo{
		Name:     "server",
		BootTime: gtype.DateTime(time.Now()),
		Version:  "1.0.1.0",
		Remark:   "XXX服务",
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Service) CanRestart(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(s.canRestart())
}

func (s *Service) CanRestartDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "后台服务")
	function := catalog.AddFunction(method, uri, "是否可在线重启")
	function.SetNote("判断当前服务是否可以在线重启")
	function.SetOutputDataExample(true)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Service) Restart(ctx gtype.Context, ps gtype.Params) {
	if !s.canRestart() {
		ctx.Error(gtype.ErrNotSupport, "当前服服不支持在线重启")
		return
	}

	s.restart(ctx)
}

func (s *Service) RestartDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "后台服务")
	function := catalog.AddFunction(method, uri, "重启服务")
	function.SetNote("重新启动当前服务")
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrNotSupport)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Service) CanUpdate(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(s.canUpdate())
}

func (s *Service) CanUpdateDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "后台服务")
	function := catalog.AddFunction(method, uri, "是否可在线更新")
	function.SetNote("判断当前服务是否可以在线更新")
	function.SetOutputDataExample(true)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Service) Update(ctx gtype.Context, ps gtype.Params) {
	s.update(ctx)
}

func (s *Service) UpdateDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	_, fileName := filepath.Split(s.cfg.Module.Path)
	note := fmt.Sprintf("安装包(必须包含文件'%s')", fileName)

	catalog := s.createCatalog(doc, "后台服务")
	function := catalog.AddFunction(method, uri, "更新服务")
	function.SetNote("上传并更新当前服务")
	function.AddInputHeader(true, "content-type", "内容类型", gtype.ContentTypeFormData)
	function.AddInputForm(true, "file", note, gtype.FormValueKindFile, nil)
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrNotSupport)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Service) extractUploadFile(ctx gtype.Context) (string, string, bool) {
	r := ctx.Request()
	uploadFile, _, err := r.FormFile("file")
	if err != nil {
		ctx.Error(gtype.ErrInput, "invalid file: ", err)
		return "", "", false
	}
	defer uploadFile.Close()

	buf := &bytes.Buffer{}
	fileSize, err := buf.ReadFrom(uploadFile)
	if err != nil {
		ctx.Error(gtype.ErrInput, "read file error: ", err)
		return "", "", false
	}
	if fileSize < 1 {
		ctx.Error(gtype.ErrInput, "invalid file: size is zero")
		return "", "", false
	}

	binFileFolder, oldBinFileName := filepath.Split(s.cfg.Module.Path)
	tempFolder := filepath.Join(binFileFolder, ctx.NewGuid())
	err = os.MkdirAll(tempFolder, 0777)
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("create temp folder '%s' error:", tempFolder), err)
		return "", "", false
	}

	fileData := buf.Bytes()
	zipFile := &gfile.Zip{}
	err = zipFile.DecompressMemory(fileData, tempFolder)
	if err != nil {
		tarFile := &gfile.Tar{}
		err = tarFile.DecompressMemory(fileData, tempFolder)
		if err != nil {
			ctx.Error(gtype.ErrInternal, "decompress file error: ", err)
			return "", tempFolder, false
		}
	}

	newBinFilePath, err := s.getBinFilePath(tempFolder, oldBinFileName)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return "", tempFolder, false
	}
	module := &gtype.Module{Path: newBinFilePath}
	moduleName := module.Name()
	if moduleName != s.cfg.Module.Name {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("模块名称(%s)无效", moduleName))
		return "", tempFolder, false
	}
	moduleType := module.Type()
	if moduleType != s.cfg.Module.Type {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("模块名称(%s)无效", moduleType))
		return "", tempFolder, false
	}

	return newBinFilePath, tempFolder, true
}

func (s *Service) getBinFilePath(folderPath, fileName string) (string, error) {
	paths, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return "", err
	}

	for _, path := range paths {
		if path.IsDir() {
			appPath, err := s.getBinFilePath(filepath.Join(folderPath, path.Name()), fileName)
			if err != nil {
				continue
			}
			return appPath, nil
		} else {
			if path.Name() == fileName {
				return filepath.Join(folderPath, path.Name()), nil
			}
		}
	}

	return "", fmt.Errorf("服务主程序(%s)不存在", fileName)
}

func (s *Service) copyFile(source, dest string) (int64, error) {
	sourceFile, err := os.Open(source)
	if err != nil {
		return 0, err
	}
	defer sourceFile.Close()

	sourceFileInfo, err := sourceFile.Stat()
	if err != nil {
		return 0, err
	}

	destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceFileInfo.Mode())
	if err != nil {
		return 0, err
	}
	defer destFile.Close()

	return io.Copy(destFile, sourceFile)
}
