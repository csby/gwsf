package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gfile"
	"github.com/csby/gwsf/gtype"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Site struct {
	controller

	apps map[string]*gcfg.SiteApp
}

func NewSite(log gtype.Log, cfg *gcfg.Config, db gtype.TokenDatabase, chs gtype.SocketChannelCollection, docPath, optPath string) *Site {
	instance := &Site{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.dbToken = db
	instance.wsChannels = chs
	instance.apps = make(map[string]*gcfg.SiteApp)

	if cfg != nil {
		if cfg.Site.Doc.Enabled {
			instance.apps[gtype.NewGuid()] = &gcfg.SiteApp{
				Name: "文档网站",
				Path: cfg.Site.Doc.Path,
				Uri:  docPath,
			}
		}

		instance.apps[gtype.NewGuid()] = &gcfg.SiteApp{
			Name: "管理网站",
			Path: cfg.Site.Opt.Path,
			Uri:  optPath,
		}

		appCount := len(cfg.Site.Apps)
		for appIndex := 0; appIndex < appCount; appIndex++ {
			app := cfg.Site.Apps[appIndex]
			instance.apps[gtype.NewGuid()] = &gcfg.SiteApp{
				Name: app.Name,
				Path: app.Path,
				Uri:  app.Uri,
			}
		}
	}

	return instance
}

func (s *Site) GetRootFiles(ctx gtype.Context, ps gtype.Params) {
	data := make([]*gtype.SiteFile, 0)

	files, err := ioutil.ReadDir(s.cfg.Site.Root.Path)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		item := &gtype.SiteFile{
			Name:       file.Name(),
			UploadTime: gtype.DateTime(file.ModTime()),
			Url:        fmt.Sprintf("%s://%s/%s", ctx.Schema(), ctx.Host(), url.PathEscape(file.Name())),
		}

		data = append(data, item)
	}

	ctx.Success(data)
}

func (s *Site) GetRootFilesDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "网站管理", "默认站点")
	function := catalog.AddFunction(method, uri, "获取文件列表")
	function.SetNote("获取默认站点(根目录)所有文件列表，但不包括文件夹")
	function.SetOutputDataExample([]gtype.SiteFile{
		{
			Url:        "http://192.168.1.1:8080/test.txt",
			Name:       "test.txt",
			UploadTime: gtype.DateTime(time.Now()),
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Site) UploadRootFile(ctx gtype.Context, ps gtype.Params) {
	r := ctx.Request()
	uploadFile, head, err := r.FormFile("file")
	if err != nil {
		ctx.Error(gtype.ErrInput, "file invalid: ", err)
		return
	}
	defer uploadFile.Close()
	buf := &bytes.Buffer{}
	fileSize, err := buf.ReadFrom(uploadFile)
	if err != nil {
		ctx.Error(gtype.ErrInput, "read file error: ", err)
		return
	}
	if fileSize < 1 {
		ctx.Error(gtype.ErrInput, "file size is zero")
		return
	}

	folderPath := s.cfg.Site.Root.Path
	err = os.MkdirAll(folderPath, 0777)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "create folder for root error: ", err)
		return
	}

	filePath := filepath.Join(folderPath, head.Filename)
	fileWriter, err := os.Create(filePath)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "create file in root folder error: ", err)
		return
	}
	defer fileWriter.Close()

	uploadFile.Seek(0, 0)
	_, err = io.Copy(fileWriter, uploadFile)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "save file in root folder error: ", err)
		return
	}

	data := &gtype.SiteFile{
		Name:       head.Filename,
		UploadTime: gtype.DateTime(time.Now()),
		Url:        fmt.Sprintf("%s://%s/%s", ctx.Schema(), ctx.Host(), url.PathEscape(head.Filename)),
	}

	ctx.Success(data)

	s.writeWebSocketMessage(ctx.Token(), gtype.WSRootSiteUploadFile, data)
}

func (s *Site) UploadRootFileDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "网站管理", "默认站点")
	function := catalog.AddFunction(method, uri, "上传文件")
	function.SetNote("上传文件到默认站点所在根目录，成功返回已上传文件信息")
	function.AddInputHeader(true, "content-type", "内容类型", gtype.ContentTypeFormData)
	function.AddInputForm(true, "file", "网站打包文件(.zip或.tar.gz)", gtype.FormValueKindFile, nil)
	function.SetOutputDataExample(gtype.SiteFile{
		Url:        "http://192.168.1.1:8080/test.txt",
		Name:       "test.txt",
		UploadTime: gtype.DateTime(time.Now()),
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Site) DeleteRootFile(ctx gtype.Context, ps gtype.Params) {
	argument := &gtype.SiteFileFilter{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInput, "文件名称(name)文空")
		return
	}

	filePath := filepath.Join(s.cfg.Site.Root.Path, argument.Name)
	err = os.Remove(filePath)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	ctx.Success(argument)

	s.writeWebSocketMessage(ctx.Token(), gtype.WSRootSiteDeleteFile, argument)
}

func (s *Site) DeleteRootFileDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "网站管理", "默认站点")
	function := catalog.AddFunction(method, uri, "删除文件")
	function.SetNote("删除根站点所在目录的文件")
	function.SetInputJsonExample(&gtype.SiteFileFilter{
		Name: "test.txt",
	})
	function.SetOutputDataExample(gtype.SiteFileFilter{
		Name: "test.txt",
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
}

func (s *Site) GetApps(ctx gtype.Context, ps gtype.Params) {
	apps := make([]string, 0)
	for k := range s.apps {
		apps = append(apps, k)
	}
	ctx.Success(apps)
}

func (s *Site) GetAppsDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "网站管理", "应用网站")
	function := catalog.AddFunction(method, uri, "获取网站列表")
	function.SetNote("获取所有的应用网站ID")
	function.SetOutputDataExample([]string{
		gtype.NewGuid(),
		gtype.NewGuid(),
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Site) GetAppInfo(ctx gtype.Context, ps gtype.Params) {
	filter := &gtype.WebAppId{}
	err := ctx.GetJson(filter)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(filter.Id) <= 0 {
		ctx.Error(gtype.ErrInput, "id is empty")
		return
	}

	v, ok := s.apps[filter.Id]
	if !ok {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("id '%s' not exist", filter.Id))
		return
	}
	info := s.newAppInfo(filter.Id, v, ctx)

	ctx.Success(info)
}

func (s *Site) GetAppInfoDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "网站管理", "应用网站")
	function := catalog.AddFunction(method, uri, "获取网站信息")
	function.SetNote("获取指定网站网站详细信息")
	now := gtype.DateTime(time.Now())
	function.SetInputJsonExample(&gtype.WebAppId{
		Id: gtype.NewGuid(),
	})
	function.SetOutputDataExample(&gtype.WebApp{
		WebAppId:   gtype.WebAppId{Id: gtype.NewGuid()},
		Name:       "管理网站",
		Url:        "https://example.com/opt",
		Version:    "1.0.1.0",
		DeployTime: &now,
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Site) UploadApp(ctx gtype.Context, ps gtype.Params) {
	id := strings.TrimSpace(strings.ToLower(ctx.Request().FormValue("id")))
	if len(id) < 1 {
		ctx.Error(gtype.ErrInput, "标识ID(id)为空")
		return
	}
	app, ok := s.apps[id]
	if !ok {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("标识ID '%s' 不存在", id))
		return
	}
	appFolder := app.Path
	if len(appFolder) < 1 {
		ctx.Error(gtype.ErrInternal, "网站物理根路径为空")
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
		ctx.Error(gtype.ErrInput, "read file error: ", err)
		return
	}
	if fileSize < 1 {
		ctx.Error(gtype.ErrInput, "invalid file: size is zero")
		return
	}

	tempFolder := filepath.Join(filepath.Dir(appFolder), ctx.NewGuid())
	err = os.MkdirAll(tempFolder, 0777)
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("create temp folder '%s' error:", tempFolder), err)
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
			ctx.Error(gtype.ErrInternal, "decompress file error: ", err)
			return
		}
	}

	appInfo := s.newAppInfo(id, app, ctx)
	if len(appInfo.Guid) > 0 {
		guid, _, _ := s.getIdVersion(tempFolder)
		if guid != appInfo.Guid {
			ctx.Error(gtype.ErrInput, fmt.Sprintf("上传的网站(guid=%s)不匹配，期望(guid=%s)", guid, appInfo.Guid))
			return
		}
	}

	err = os.RemoveAll(appFolder)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "remove original site error: ", err)
		return
	}
	os.MkdirAll(filepath.Dir(appFolder), 0777)
	err = os.Rename(tempFolder, appFolder)
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("rename folder '%s' error:", appFolder), err)
		return
	}

	appInfo = s.newAppInfo(id, app, ctx)
	ctx.Success(appInfo)

	s.writeWebSocketMessage(ctx.Token(), gtype.WSSiteUpload, appInfo)
}

func (s *Site) UploadAppDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "网站管理", "应用网站")
	function := catalog.AddFunction(method, uri, "上传网站")
	function.SetNote("上传网站打包文件(.zip或.tar.gz)，并替换之前已发布的网站")
	now := gtype.DateTime(time.Now())
	function.AddInputHeader(true, "content-type", "内容类型", gtype.ContentTypeFormData)
	function.AddInputForm(true, "id", "标识ID", gtype.FormValueKindText, "")
	function.AddInputForm(true, "file", "网站打包文件(.zip或.tar.gz)", gtype.FormValueKindFile, nil)
	function.SetOutputDataExample(&gtype.WebApp{
		WebAppId:   gtype.WebAppId{Id: gtype.NewGuid()},
		Name:       "管理网站",
		Url:        "https://example.com/opt",
		Version:    "1.0.1.0",
		DeployTime: &now,
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Site) newAppInfo(id string, app *gcfg.SiteApp, ctx gtype.Context) *gtype.WebApp {
	info := &gtype.WebApp{
		Name: app.Name,
		Url:  fmt.Sprintf("%s://%s%s/", ctx.Schema(), ctx.Host(), app.Uri),
		Root: app.Path,
	}
	info.Id = id

	fi, err := os.Stat(app.Path)
	if os.IsNotExist(err) == false {
		if fi.IsDir() {
			deployTime := gtype.DateTime(fi.ModTime())
			info.DeployTime = &deployTime
			info.Guid, info.Version, _ = s.getIdVersion(app.Path)
		}
	}

	return info
}

func (s *Site) getIdVersion(folderPath string) (string, string, error) {
	/*
		{
		  "guid": "FDC3BDEFB79CEC8EB8211D2499E04704",
		  "version": "1.0.1.1"
		}
	*/
	filePath := filepath.Join(folderPath, "site.json")
	fi, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return "", "", err
	}
	if fi.IsDir() {
		return "", "", fmt.Errorf("%s is not file", filePath)
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", "", err
	}
	app := &gtype.WebApp{}
	err = json.Unmarshal(data, app)
	if err != nil {
		return "", "", err
	}

	return app.Guid, app.Version, nil
}
