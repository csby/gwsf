package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/csby/gwsf/gfile"
	"github.com/csby/gwsf/gmodel"
	"github.com/csby/gwsf/gtype"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func (s *Service) GetCustomShellInfo(ctx gtype.Context, ps gtype.Params) {
	executable := filepath.Join(filepath.Dir(s.cfg.Module.Path), "gshell")
	if runtime.GOOS == "windows" {
		executable += ".exe"
	}

	info, err := s.getCustomShellInfo(executable)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	ctx.Success(info)
}

func (s *Service) GetCustomShellInfoDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
	function := catalog.AddFunction(method, uri, "获取服务外壳程序信息")
	function.SetOutputDataExample([]*gmodel.ServiceCustomShellInfo{
		{
			Name:       "gshell",
			Version:    "1.0.1.0",
			DeployTime: gtype.DateTime(time.Now()),
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
}

func (s *Service) UpdateCustomShell(ctx gtype.Context, ps gtype.Params) {
	runningServices := make([]string, 0)
	fileName := "gshell"
	if runtime.GOOS == "windows" {
		fileName += ".exe"

		services := s.getCustoms()
		c := len(services)
		for i := 0; i < c; i++ {
			item := services[i]
			if item == nil {
				continue
			}
			serviceName := item.GetServiceName()
			item.Status, _ = s.getStatus(serviceName)
			if item.Status == gtype.ServerStatusRunning {
				runningServices = append(runningServices, serviceName)
			}
		}
	}
	newBinFilePath, folder, ok := s.extractCustomShellFile(ctx, fileName)
	if !ok {
		if len(folder) > 0 {
			os.RemoveAll(folder)
		}
		return
	}
	defer os.RemoveAll(folder)

	// stop the service
	rc := len(runningServices)
	if rc > 0 {
		for ri := 0; ri < rc; ri++ {
			s.stop(runningServices[ri])
		}
		time.Sleep(time.Second)
	}

	// start the service
	defer func(names []string) {
		c := len(names)
		for i := 0; i < c; i++ {
			go s.start(names[i])
		}
	}(runningServices)

	oldBinFilePath := filepath.Join(filepath.Dir(s.cfg.Module.Path), fileName)
	err := os.Remove(oldBinFilePath)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	_, err = s.copyFile(newBinFilePath, oldBinFilePath)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	ctx.Success(nil)
}

func (s *Service) UpdateCustomShellDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	fileName := "gshell"
	if runtime.GOOS == "windows" {
		fileName += ".exe"
	}
	note := fmt.Sprintf("压缩包(必须包含文件'%s')", fileName)

	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogCustom)
	function := catalog.AddFunction(method, uri, "更新服务外壳程序")
	function.SetNote("上传并更新服务外壳程序")
	function.AddInputHeader(true, "content-type", "内容类型", gtype.ContentTypeFormData)
	function.AddInputForm(true, "file", note, gtype.FormValueKindFile, nil)
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrNotSupport)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Service) getCustomShellInfo(path string) (*gmodel.ServiceCustomShellInfo, error) {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("%s is not file", path)
	}

	buf := &bytes.Buffer{}
	cmd := exec.Command(path, "-V")
	cmd.Stdout = buf
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	bufData := buf.Bytes()
	if len(bufData) < 1 {
		return nil, fmt.Errorf("无效的程序")
	}
	jsonData, err := base64.StdEncoding.DecodeString(string(bufData))
	if err != nil {
		return nil, err
	}
	info := &gmodel.ServiceCustomShellInfo{}
	err = json.Unmarshal(jsonData, info)
	if err != nil {
		return nil, err
	}
	info.DeployTime = gtype.DateTime(fi.ModTime())

	return info, nil
}

func (s *Service) extractCustomShellFile(ctx gtype.Context, fileName string) (string, string, bool) {
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

	binFileFolder, _ := filepath.Split(s.cfg.Module.Path)
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

	newBinFilePath, err := s.getBinFilePath(tempFolder, fileName)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return "", tempFolder, false
	}
	info, err := s.getCustomShellInfo(newBinFilePath)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return "", tempFolder, false
	}

	if info.Name != "gshell" {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("模块名称(%s)无效", info.Name))
		return "", tempFolder, false
	}

	return newBinFilePath, tempFolder, true
}
