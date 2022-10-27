package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gfile"
	"github.com/csby/gwsf/gmodel"
	"github.com/csby/gwsf/gtype"
	"github.com/kardianos/service"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	svcCatalogRoot   = "系统服务"
	svcCatalogTomcat = "tomcat"
	svcCatalogNginx  = "nginx"
	svcCatalogCustom = "自定义"
	svcCatalogOther  = "其他"
	svcCatalogFile   = "文件"
)

type Service struct {
	controller

	svcMgr      gtype.SvcUpdMgr
	bootTime    time.Time
	fileServers []*ServiceFileServer
}

func NewService(log gtype.Log, cfg *gcfg.Config, svcMgr gtype.SvcUpdMgr, wsc gtype.SocketChannelCollection) *Service {
	inst := &Service{}
	inst.SetLog(log)
	inst.cfg = cfg
	inst.svcMgr = svcMgr
	inst.wsChannels = wsc

	inst.fileServers = make([]*ServiceFileServer, 0)
	if cfg != nil {
		inst.bootTime = cfg.Svc.BootTime
		c := len(cfg.Sys.Svc.Files)
		for i := 0; i < c; i++ {
			fs := cfg.Sys.Svc.Files[i]
			if fs == nil {
				continue
			}
			if len(fs.Path) < 1 {
				continue
			}
			if len(fs.Root) < 1 {
				continue
			}

			inst.fileServers = append(inst.fileServers, &ServiceFileServer{
				Root:    fs.Root,
				Path:    fs.Path,
				Enabled: fs.Enabled,
			})
		}
	} else {
		inst.bootTime = time.Now()
	}

	return inst
}

func (s *Service) FileServers() []*ServiceFileServer {
	return s.fileServers
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

		if len(cfg.Svc.Name) > 0 {
			data.Name = cfg.Svc.Name
		}
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

	s.doRestart(ctx)
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

func (s *Service) getStatus(name string) (gtype.ServerStatus, error) {
	cfg := &service.Config{}
	cfg.Name = name
	//if runtime.GOOS == "linux" {
	//	cfg.Name = fmt.Sprintf("%s.service", name)
	//}

	svc, err := service.New(nil, cfg)
	if err != nil {
		return gtype.ServerStatusUnknown, err
	}

	status, err := svc.Status()
	if err != nil {
		if err == service.ErrNotInstalled || err.Error() == "the service is not installed" {
			if strings.Contains(name, "@") {
				return gtype.ServerStatusStopped, nil
			}
			return gtype.ServerStatusUnknown, nil
		} else if err.Error() == "service in failed state" {
			return gtype.ServerStatusStopped, nil
		} else if err.Error() == "the service is not installed" {
			return gtype.ServerStatusStopped, nil
		} else if strings.Contains(name, "@") {
			return gtype.ServerStatusStopped, nil
		}
		return gtype.ServerStatusUnknown, err
	}

	return gtype.ServerStatus(status), nil
}

func (s *Service) uninstall(name string) error {
	cfg := &service.Config{Name: name}
	svc, err := service.New(nil, cfg)
	if err != nil {
		return err
	}

	return svc.Uninstall()
}

func (s *Service) start(name string) error {
	cfg := &service.Config{Name: name}
	svc, err := service.New(nil, cfg)
	if err != nil {
		return err
	}

	return svc.Start()
}

func (s *Service) stop(name string) error {
	cfg := &service.Config{Name: name}
	svc, err := service.New(nil, cfg)
	if err != nil {
		return err
	}

	return svc.Stop()
}

func (s *Service) restart(name string) error {
	cfg := &service.Config{Name: name}
	svc, err := service.New(nil, cfg)
	if err != nil {
		return err
	}

	return svc.Restart()
}

func (s *Service) getFiles(root, folder string) gmodel.ServiceLogFileCollection {
	files := make(gmodel.ServiceLogFileCollection, 0)

	fs, fe := ioutil.ReadDir(filepath.Join(root, folder))
	if fe != nil {
		return files
	}

	for _, f := range fs {
		if f.IsDir() {
			continue
		}

		file := &gmodel.ServiceLogFile{}
		files = append(files, file)
		file.Name = f.Name()
		file.Size = f.Size()
		file.ModTime = gtype.DateTime(f.ModTime())
		file.SizeText = s.sizeToText(float64(file.Size))
		file.Path = base64.URLEncoding.EncodeToString([]byte(filepath.Join(folder, file.Name)))
	}

	return files
}

func (s *Service) removeFiles(folder string, now time.Time, expiration time.Duration) {
	fs, fe := ioutil.ReadDir(folder)
	if fe != nil {
		return
	}

	for _, f := range fs {
		path := filepath.Join(folder, f.Name())
		if f.IsDir() {
			s.removeFiles(path, now, expiration)
		} else {
			if now.Sub(f.ModTime()) > expiration {
				os.Remove(path)
			}
		}
	}
}

func (s *Service) getJsonVersion(folderPath string) (string, error) {
	/*
		{
		  "version": "1.0.1.1"
		}
	*/
	filePath := filepath.Join(folderPath, "version.json")
	fi, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return "", err
	}
	if fi.IsDir() {
		return "", fmt.Errorf("%s is not file", filePath)
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	app := &gtype.WebApp{}
	err = json.Unmarshal(data, app)
	if err != nil {
		return "", err
	}

	return app.Version, nil
}

func (s *Service) getTextVersion(folderPath string) (string, error) {
	/*
		1.0.1.1
	*/
	filePath := filepath.Join(folderPath, "version.txt")
	fi, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return "", err
	}
	if fi.IsDir() {
		return "", fmt.Errorf("%s is not file", filePath)
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (s *Service) getFileInfos(cfg *gmodel.FileInfo, root string) {
	if cfg == nil {
		return
	}
	if cfg.Children == nil {
		cfg.Children = make(gmodel.FileInfoCollection, 0)
	}

	if len(root) < 1 {
		return
	}
	fi, ie := os.Stat(root)
	if os.IsNotExist(ie) {
		return
	}
	if !fi.IsDir() {
		return
	}
	cfg.Name = fi.Name()
	cfg.Folder = true
	cfg.Size = 0
	cfg.Time = gtype.DateTime(fi.ModTime())
	cfg.Path = base64.URLEncoding.EncodeToString([]byte(root))

	s.getFileInfo(cfg, root, "")
}

func (s *Service) getFileInfo(cfg *gmodel.FileInfo, root, folder string) {
	if cfg == nil {
		return
	}
	if cfg.Children == nil {
		cfg.Children = make(gmodel.FileInfoCollection, 0)
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
		item := &gmodel.FileInfo{
			Name:     name,
			Path:     base64.URLEncoding.EncodeToString([]byte(filepath.Join(root, path))),
			Time:     gtype.DateTime(f.ModTime()),
			Children: make(gmodel.FileInfoCollection, 0),
		}
		cfg.Children = append(cfg.Children, item)
		if f.IsDir() {
			item.Folder = true
			item.Size = 0
			s.getFileInfo(item, root, path)
		} else {
			item.Size = f.Size()
		}
		cfg.Size += item.Size
	}
}

func (s *Service) getLogFiles(cfg *gmodel.ServiceLogFile, root, folder string) {
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
			s.getLogFiles(item, root, path)
		} else {
			item.Size = f.Size()
			item.SizeText = s.sizeToText(float64(item.Size))
		}
	}
}
