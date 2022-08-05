package controller

import (
	"github.com/csby/gwsf/gtype"
	"io"
	"os"
	"path/filepath"
)

const (
	maxMemory = 1024 << 20 // 1024 MB
)

type ServiceFileServer struct {
	Root    string
	Path    string
	Enabled bool
}

func (s *ServiceFileServer) Upload(ctx gtype.Context, ps gtype.Params) {
	request := ctx.Request()
	err := request.ParseMultipartForm(maxMemory)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	fileCount := len(request.MultipartForm.File)
	if fileCount < 1 {
		ctx.Error(gtype.ErrInput, "not have any file")
		return
	}

	for _, files := range request.MultipartForm.File {
		if len(files) != 1 {
			ctx.Error(gtype.ErrInput, "too many files")
			return
		}
		file := files[0]
		f, e := file.Open()
		if e != nil {
			ctx.Error(gtype.ErrInternal, e)
			return
		}

		targetFilePath := filepath.Join(s.Root, file.Filename)
		err = s.saveFile(targetFilePath, f)
		if err != nil {
			ctx.Error(gtype.ErrInternal, err)
			return
		}
	}

	ctx.Success(nil)
}

func (s *ServiceFileServer) saveFile(path string, file io.Reader) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}

	writer, err := os.Create(path)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, file)

	return err
}
