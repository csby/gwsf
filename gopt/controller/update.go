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

type Update struct {
	controller

	svcMgr gtype.SvcUpdMgr
}

func NewUpdate(log gtype.Log, cfg *gcfg.Config, svcMgr gtype.SvcUpdMgr) *Update {
	instance := &Update{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.svcMgr = svcMgr

	return instance
}

func (s *Update) Enable(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(s.isEnable())
}

func (s *Update) EnableDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "更新管理")
	function := catalog.AddFunction(method, uri, "是否支持")
	function.SetNote("判断当前服务是否支持更新管理，当后台服务运行在Windows下时为true，其它为false")
	function.SetOutputDataExample(false)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Update) Info(ctx gtype.Context, ps gtype.Params) {
	data, err := s.info()
	if err != nil {
		ctx.Error(err)
	}

	ctx.Success(data)
}

func (s *Update) InfoDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	bootTime := gtype.DateTime(time.Now())
	catalog := s.createCatalog(doc, "更新管理")
	function := catalog.AddFunction(method, uri, "获取服务信息")
	function.SetNote("获取当前服务信息")
	function.SetOutputDataExample(&gtype.SvcUpdInfo{
		Name:     "server",
		BootTime: &bootTime,
		Version:  "1.0.1.0",
		Remark:   "XXX服务",
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Update) CanRestart(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(s.canRestart())
}

func (s *Update) CanRestartDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "更新管理")
	function := catalog.AddFunction(method, uri, "是否可在线重启")
	function.SetNote("判断当前服务是否可以在线重启")
	function.SetOutputDataExample(false)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Update) Restart(ctx gtype.Context, ps gtype.Params) {
	if !s.canRestart() {
		ctx.Error(gtype.ErrNotSupport)
		return
	}

	err := s.restart()
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Success(nil)
}

func (s *Update) RestartDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "更新管理")
	function := catalog.AddFunction(method, uri, "重启服务")
	function.SetNote("重新启动当前服务")
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrNotSupport)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Update) CanUpdate(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(s.canUpdate())
}

func (s *Update) CanUpdateDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "更新管理")
	function := catalog.AddFunction(method, uri, "是否可在线更新")
	function.SetNote("判断当前服务是否可以在线更新")
	function.SetOutputDataExample(false)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Update) Update(ctx gtype.Context, ps gtype.Params) {
	err := s.update(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Success(nil)
}

func (s *Update) UpdateDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	fileName := s.executeFileName()
	note := fmt.Sprintf("安装包(必须包含文件'%s')", fileName)

	catalog := s.createCatalog(doc, "更新管理")
	function := catalog.AddFunction(method, uri, "更新服务")
	function.SetNote("上传并更新当前服务")
	function.AddInputHeader(true, "content-type", "内容类型", gtype.ContentTypeFormData)
	function.AddInputForm(true, "file", note, gtype.FormValueKindFile, nil)
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrNotSupport)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Update) serviceName() string {
	return "gwsfupd"
}

func (s *Update) getBinFilePath(folderPath, fileName string) (string, error) {
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

func (s *Update) extractUploadFile(ctx gtype.Context) (string, string, gtype.Error) {
	r := ctx.Request()
	uploadFile, _, err := r.FormFile("file")
	if err != nil {
		return "", "", gtype.ErrInput.SetDetail("invalid file: ", err)
	}
	defer uploadFile.Close()

	buf := &bytes.Buffer{}
	fileSize, err := buf.ReadFrom(uploadFile)
	if err != nil {
		return "", "", gtype.ErrInput.SetDetail("read file error: ", err)
	}
	if fileSize < 1 {
		ctx.Error(gtype.ErrInput, "invalid file: size is zero")
		return "", "", gtype.ErrInput.SetDetail("invalid file: size is zero")
	}

	binFileFolder, _ := filepath.Split(s.cfg.Module.Path)
	tempFolder := filepath.Join(binFileFolder, ctx.NewGuid())
	err = os.MkdirAll(tempFolder, 0777)
	if err != nil {
		return "", "", gtype.ErrInternal.SetDetail(fmt.Sprintf("create temp folder '%s' error:", tempFolder), err)
	}

	fileData := buf.Bytes()
	zipFile := &gfile.Zip{}
	err = zipFile.DecompressMemory(fileData, tempFolder)
	if err != nil {
		tarFile := &gfile.Tar{}
		err = tarFile.DecompressMemory(fileData, tempFolder)
		if err != nil {
			return "", tempFolder, gtype.ErrInternal.SetDetail("decompress file error: ", err)
		}
	}

	oldBinFileName := s.executeFileName()
	newBinFilePath, err := s.getBinFilePath(tempFolder, oldBinFileName)
	if err != nil {
		return "", tempFolder, gtype.ErrInput.SetDetail(err)
	}
	module := &gtype.Module{Path: newBinFilePath}
	moduleName := module.Name()
	if moduleName != gtype.SvcUpdModuleName {
		return "", tempFolder, gtype.ErrInput.SetDetail(fmt.Sprintf("模块名称(%s)无效", moduleName))
	}
	moduleType := module.Type()
	if moduleType != gtype.SvcUpdModuleType {
		return "", tempFolder, gtype.ErrInput.SetDetail(fmt.Sprintf("模块名称(%s)无效", moduleType))
	}

	return newBinFilePath, tempFolder, nil
}

func (s *Update) copyFile(source, dest string) (int64, error) {
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
