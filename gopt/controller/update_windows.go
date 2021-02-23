package controller

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
	"os"
	"path/filepath"
	"strings"
)

func (s *Update) isEnable() bool {
	return true
}

func (s *Update) info() (*gtype.SvcUpdInfo, gtype.Error) {
	data := &gtype.SvcUpdInfo{}
	data.Name = s.serviceName()
	data.Status = 0

	info, err := s.svcMgr.RemoteInfo()
	if err == nil {
		data.Version = info.Version
		data.Remark = info.Remark
		data.BootTime = info.BootTime
		data.Status = 2
	} else {
		status, err := s.svcMgr.Status(data.Name)
		if err == nil {
			if status == gtype.ServerStatusStopped {
				data.Status = 1
			} else if status == gtype.ServerStatusRunning {
				data.Status = 2
			}
		}
	}

	return data, nil
}

func (s *Update) canRestart() bool {
	_, err := s.svcMgr.Status(s.serviceName())
	if err != nil {
		return false
	} else {
		return true
	}
}

func (s *Update) restart() gtype.Error {
	status, err := s.svcMgr.Status(s.serviceName())
	if err != nil {
		return gtype.ErrInternal.SetDetail(err)
	}

	if status == gtype.ServerStatusRunning {
		err = s.svcMgr.Restart(s.serviceName())
	} else {
		err = s.svcMgr.Start(s.serviceName())
	}

	if err != nil {
		return gtype.ErrInternal.SetDetail(err)
	}

	return nil
}

func (s *Update) canUpdate() bool {
	info, err := s.svcMgr.RemoteInfo()
	if err == nil {
		if info.Interactive {
			return false
		}
	}

	return true
}

func (s *Update) update(ctx gtype.Context) gtype.Error {
	newBinFilePath, folder, e := s.extractUploadFile(ctx)
	if e != nil {
		if len(folder) > 0 {
			os.RemoveAll(folder)
		}
		return e
	}
	defer os.RemoveAll(folder)

	if !s.canUpdate() {
		return gtype.ErrNotSupport.SetDetail("服务不支持在线更新")
	}

	svcName := s.serviceName()
	status, err := s.svcMgr.Status(svcName)
	if err != nil {
		if strings.Contains(err.Error(), "Access is denied") {
			return gtype.ErrInternal.SetDetail(err)
		}

		binFileFolder, _ := filepath.Split(s.cfg.Module.Path)
		binFilePath := filepath.Join(binFileFolder, s.executeFileName())
		_, err := s.copyFile(newBinFilePath, binFilePath)
		if err != nil {
			return gtype.ErrInternal.SetDetail(fmt.Errorf("copy file error: %v", err))
		}

		err = s.svcMgr.Install(svcName, binFilePath)
		if err != nil {
			return gtype.ErrInternal.SetDetail(fmt.Errorf("install service '%s' error: %v", svcName, err))
		}

		err = s.svcMgr.Start(svcName)
		if err != nil {
			return gtype.ErrInternal.SetDetail(fmt.Errorf("start service '%s' error: %v", svcName, err))
		}
	} else {
		if status != gtype.ServerStatusRunning {
			err = s.svcMgr.Start(svcName)
			if err != nil {
				return gtype.ErrInternal.SetDetail(fmt.Errorf("stop service '%s' error: %v", svcName, err))
			}
		}

		info, err := s.svcMgr.RemoteInfo()
		if err != nil {
			return gtype.ErrInternal.SetDetail(fmt.Errorf("get service '%s' info error: %v", svcName, err))
		}

		err = s.svcMgr.Stop(svcName)
		if err != nil {
			return gtype.ErrInternal.SetDetail(fmt.Errorf("stop service '%s' error: %v", svcName, err))
		}

		err = os.Remove(info.Path)
		if err != nil {
			return gtype.ErrInternal.SetDetail(fmt.Errorf("remove service '%s' execute file error: %v", svcName, err))
		}

		_, err = s.copyFile(newBinFilePath, info.Path)
		if err != nil {
			return gtype.ErrInternal.SetDetail(fmt.Errorf("copy file error: %v", err))
		}

		err = s.svcMgr.Start(svcName)
		if err != nil {
			return gtype.ErrInternal.SetDetail(fmt.Errorf("start service '%s' error: %v", svcName, err))
		}
	}

	return nil
}

func (s *Update) executeFileName() string {
	return "gwsfupd.exe"
}
