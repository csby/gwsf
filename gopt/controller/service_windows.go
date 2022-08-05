package controller

import (
	"github.com/csby/gwsf/gtype"
	"os"
)

func (s *Service) canRestart() bool {
	if s.cfg == nil {
		return false
	}

	if s.cfg.Svc.Restart == nil {
		return false
	}

	return true
}

func (s *Service) doRestart(ctx gtype.Context) {
	err := s.svcMgr.RemoteRestart(s.cfg.Svc.Name)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	ctx.Success(nil)
}

func (s *Service) canUpdate() bool {
	return s.canRestart()
}

func (s *Service) update(ctx gtype.Context) {
	newBinFilePath, folder, ok := s.extractUploadFile(ctx)
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

	svcName := s.cfg.Svc.Name
	svcPath := s.cfg.Module.Path
	err := s.svcMgr.RemoteUpdate(svcName, svcPath, newBinFilePath, folder)
	if err != nil {
		os.RemoveAll(folder)
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	ctx.Success(nil)
}
