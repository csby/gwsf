package controller

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/csby/gwsf/gmodel"
	"github.com/csby/gwsf/gtype"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func (s *Service) ViewFile(ctx gtype.Context, ps gtype.Params) {
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
	pathValue := string(pathData)
	fi, fe := os.Stat(pathValue)
	if os.IsNotExist(fe) {
		ctx.Error(gtype.ErrInput, fe)
		return
	}
	if fi.IsDir() {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("指定的路径(%s)为文件夹", pathValue))
		return
	}

	file, le := os.OpenFile(pathValue, os.O_RDONLY, 0666)
	if le != nil {
		ctx.Error(gtype.ErrInternal, le)
		return
	}
	defer file.Close()

	extName := strings.ToLower(path.Ext(pathValue))
	if extName == ".xml" {
		ctx.Response().Header().Set("Content-Type", "application/xml;charset=utf-8")
	} else if extName == ".json" {
		ctx.Response().Header().Set("Content-Type", "application/json;charset=utf-8")
	} else if extName == ".css" {
		ctx.Response().Header().Set("Content-Type", "text/css;charset=utf-8")
	} else if extName == ".jpeg" {
		ctx.Response().Header().Set("Content-Type", "image/jpeg")
	} else if extName == ".png" {
		ctx.Response().Header().Set("Content-Type", "image/png")
	} else {
		ctx.Response().Header().Set("Content-Type", "text/plain;charset=utf-8")
	}

	contentLength := fi.Size()
	ctx.Response().Header().Set("Content-Length", fmt.Sprint(contentLength))

	io.Copy(ctx.Response(), file)
}

func (s *Service) ViewFileDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogFile)
	function := catalog.AddFunction(method, uri, "查看文件")
	function.SetNote("返回文本内容信息")
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) DownloadFile(ctx gtype.Context, ps gtype.Params) {
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
	pathValue := string(pathData)
	fi, fe := os.Stat(pathValue)
	if os.IsNotExist(fe) {
		ctx.Error(gtype.ErrInput, fe)
		return
	}

	if fi.IsDir() {
		fileName := fmt.Sprintf("%s.zip", filepath.Base(pathValue))
		ctx.Response().Header().Set("Content-Disposition", fmt.Sprint("attachment; filename=", fileName))
		s.compressFolder(ctx.Response(), pathValue, "", nil)
	} else {
		file, le := os.OpenFile(pathValue, os.O_RDONLY, 0666)
		if le != nil {
			ctx.Error(gtype.ErrInternal, le)
			return
		}
		defer file.Close()

		fileName := filepath.Base(pathValue)
		ctx.Response().Header().Set("Content-Disposition", fmt.Sprint("attachment; filename=", fileName))

		contentLength := fi.Size()
		ctx.Response().Header().Set("Content-Length", fmt.Sprint(contentLength))

		io.Copy(ctx.Response(), file)
	}
}

func (s *Service) DownloadFileDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogFile)
	function := catalog.AddFunction(method, uri, "下载文件")
	function.SetNote("返回文件或文件夹压缩包")
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) ModFile(ctx gtype.Context, ps gtype.Params) {
	parent := strings.TrimSpace(ctx.Request().FormValue("parent"))
	if len(parent) < 1 {
		ctx.Error(gtype.ErrInput, "父级路径(parent)为空")
		return
	}
	parentData, err := base64.URLEncoding.DecodeString(parent)
	if err != nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("父级路径base64(%s)无效", parent))
		return
	}
	if len(parentData) < 1 {
		ctx.Error(gtype.ErrInput, "父级路径为空")
		return
	}
	parentValue := string(parentData)
	fi, fe := os.Stat(parentValue)
	if os.IsNotExist(fe) {
		ctx.Error(gtype.ErrInput, fe)
		return
	}
	folderPath := parentValue
	if !fi.IsDir() {
		folderPath = filepath.Dir(parentValue)
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

	ctx.Success(base64.URLEncoding.EncodeToString([]byte(targetFilePath)))
}

func (s *Service) ModFileDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogFile)
	function := catalog.AddFunction(method, uri, "上传文件")
	function.SetNote("上传应并更新文件文件, 成功时返回已上传文件的路径")
	function.AddInputHeader(true, "content-type", "内容类型", gtype.ContentTypeFormData)
	function.AddInputForm(true, "parent", "父级路径", gtype.FormValueKindText, "")
	function.AddInputForm(true, "file", "应用程序配置文件", gtype.FormValueKindFile, nil)
	function.SetOutputDataExample(gtype.NewGuid())
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) DeleteFile(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.FileInfoArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
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
	fullPath := string(pathData)
	if len(fullPath) < 1 {
		ctx.Error(gtype.ErrInput, "路径为空")
		return
	}
	err = os.RemoveAll(fullPath)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "删除失败: ", err)
		return
	}

	ctx.Success(base64.URLEncoding.EncodeToString([]byte(filepath.Dir(fullPath))))
}

func (s *Service) DeleteFileDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogFile)
	function := catalog.AddFunction(method, uri, "删除文件")
	function.SetNote("删除指定的文件或文件夹,成功时返回被删除文件或文件的目录路径")
	function.SetInputJsonExample(&gmodel.FileInfoArgument{})
	function.SetOutputDataExample(gtype.NewGuid())
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}
