package controller

import (
	"github.com/csby/gwsf/gtype"
	"os"
	"runtime"
	"time"
)

func (s *Service) canRestart() bool {
	if s.cfg == nil {
		return false
	}

	if s.cfg.Svc.Restart == nil {
		return false
	}

	if runtime.GOOS == "linux" {
		return true
	} else {
		return false
	}
}

func (s *Service) restart(ctx gtype.Context, ps gtype.Params) {
	go func() {
		time.Sleep(2 * time.Second)
		err := s.cfg.Svc.Restart()
		if err != nil {
			s.LogError("重启服务失败:", err)
		}
		os.Exit(1)
	}()

	ctx.Success(nil)
}

func (s *Service) canUpdate() bool {
	return s.canRestart()
}

func (s *Service) update(ctx gtype.Context, ps gtype.Params) {
	newBinFilePath, folder, ok := s.extractUploadFile(ctx, ps)
	if !ok {
		if len(folder) > 0 {
			os.RemoveAll(folder)
		}
		return
	}
	defer os.RemoveAll(folder)

	if !s.canUpdate() {
		ctx.Error(gtype.ErrNotSupport, "服务不支持在线更新")
		return
	}

	oldBinFilePath := s.cfg.Module.Path
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

	go func() {
		time.Sleep(2 * time.Second)
		err := s.cfg.Svc.Restart()
		if err != nil {
			s.LogError("重启服务失败:", err)
		}
		os.Exit(1)
	}()

	ctx.Success(nil)
}
