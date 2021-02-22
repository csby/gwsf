package gopt

import (
	"encoding/json"
	"fmt"
	"github.com/csby/gwsf/gclient"
	"github.com/csby/gwsf/gtype"
	"github.com/kardianos/service"
)

const (
	svcUpdMgrUrl = "http://127.0.0.1:9606"
)

type SvcUpdMgr struct {
}

func (s *SvcUpdMgr) Start(name string) error {
	cfg := &service.Config{Name: name}
	svc, err := service.New(nil, cfg)
	if err != nil {
		return err
	}

	return svc.Start()
}

func (s *SvcUpdMgr) Stop(name string) error {
	cfg := &service.Config{Name: name}
	svc, err := service.New(nil, cfg)
	if err != nil {
		return err
	}

	return svc.Stop()
}

func (s *SvcUpdMgr) Restart(name string) error {
	cfg := &service.Config{Name: name}
	svc, err := service.New(nil, cfg)
	if err != nil {
		return err
	}

	return svc.Restart()
}

func (s *SvcUpdMgr) Status(name string) (gtype.ServerStatus, error) {
	cfg := &service.Config{Name: name}
	svc, err := service.New(nil, cfg)
	if err != nil {
		return gtype.ServerStatusUnknown, err
	}

	status, err := svc.Status()
	return gtype.ServerStatus(status), err
}

func (s *SvcUpdMgr) Install(name, path string) error {
	cfg := &service.Config{
		Name:        name,
		DisplayName: name,
		Description: name,
		Executable:  path,
	}
	svc, err := service.New(nil, cfg)
	if err != nil {
		return err
	}

	return svc.Install()
}

func (s *SvcUpdMgr) RemoteInfo() (*gtype.SvcUpdResult, error) {
	httpClient := &gclient.Http{}
	argument := &gtype.SvcUpdArgs{Action: "info"}
	_, output, _, _, err := httpClient.PostJson(svcUpdMgrUrl, argument)
	if err != nil {
		return nil, err
	}

	result := &gtype.SvcUpdResult{Interactive: true}
	err = json.Unmarshal(output, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *SvcUpdMgr) RemoteRestart(name string) error {
	httpClient := &gclient.Http{}
	argument := &gtype.SvcUpdArgs{
		Action: "restart",
		Name:   name,
	}
	_, output, _, _, err := httpClient.PostJson(svcUpdMgrUrl, argument)
	if err != nil {
		return err
	}

	result := &gtype.SvcUpdResult{}
	err = json.Unmarshal(output, result)
	if err != nil {
		return err
	}

	if result.Code == 0 {
		return nil
	}
	return fmt.Errorf(result.Error)
}

func (s *SvcUpdMgr) RemoteUpdate(name, path, updateFile, updateFolder string) error {
	httpClient := &gclient.Http{}
	argument := &gtype.SvcUpdArgs{
		Action:       "update",
		Name:         name,
		Path:         path,
		UpdateFile:   updateFile,
		UpdateFolder: updateFolder,
	}
	_, output, _, _, err := httpClient.PostJson(svcUpdMgrUrl, argument)
	if err != nil {
		return err
	}

	result := &gtype.SvcUpdResult{}
	err = json.Unmarshal(output, result)
	if err != nil {
		return err
	}

	if result.Code == 0 {
		return nil
	}
	return fmt.Errorf(result.Error)
}
