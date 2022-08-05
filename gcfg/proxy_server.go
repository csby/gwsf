package gcfg

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
)

type ProxyServer struct {
	Id      string `json:"id" note:"标识ID"`
	Name    string `json:"name" note:"名称"`
	Disable bool   `json:"disable" note:"已禁用"`
	TLS     bool   `json:"tls" note:"传入是否为TLS连接"`

	IP   string `json:"ip" note:"监听地址，空表示所有IP地址"`
	Port string `json:"port" note:"监听端口"`

	Targets []*ProxyTarget `json:"targets" note:"目标地址"`
}

func (s *ProxyServer) initId() {
	if len(s.Id) < 1 {
		s.Id = gtype.NewGuid()
	}

	count := len(s.Targets)
	for i := 0; i < count; i++ {
		item := s.Targets[i]
		if item == nil {
			continue
		}
		if len(item.Id) < 1 {
			item.Id = gtype.NewGuid()
		}
		item.Alive = false
		item.ConnCount = 0

		for j := 0; j < len(item.Spares); j++ {
			spare := item.Spares[j]
			if spare == nil {
				continue
			}
			spare.Alive = false
			spare.ConnCount = 0
		}
	}
}

func (s *ProxyServer) InitAddrId() {
	c := len(s.Targets)
	for i := 0; i < c; i++ {
		item := s.Targets[i]
		if item == nil {
			continue
		}
		item.InitAddrId(s.Id)
	}
}

func (s *ProxyServer) UniqueId() string {
	return fmt.Sprintf("%s:%s", s.IP, s.Port)
}

func (s *ProxyServer) AddTarget(target *ProxyTarget) error {
	if target == nil {
		return fmt.Errorf("target is nil")
	}

	count := len(s.Targets)
	for i := 0; i < count; i++ {
		if target.Domain == s.Targets[i].Domain && target.Path == s.Targets[i].Path {
			return fmt.Errorf("domain '%s' and path '%s' has been existed", target.Domain, target.Path)
		}
	}

	target.InitAddrId(s.Id)
	s.Targets = append(s.Targets, target)

	return nil
}

func (s *ProxyServer) DeleteTarget(id string) error {
	targets := make([]*ProxyTarget, 0)
	count := len(s.Targets)
	deletedCount := 0
	for i := 0; i < count; i++ {
		if id == s.Targets[i].Id {
			deletedCount++
			continue
		}
		targets = append(targets, s.Targets[i])
	}
	if deletedCount <= 0 {
		return fmt.Errorf("target id '%s' not existed", id)
	}

	s.Targets = targets

	return nil
}

func (s *ProxyServer) ModifyTarget(target *ProxyTarget) (*ProxyTarget, error) {
	if target == nil {
		return nil, fmt.Errorf("target is nil")
	}

	count := len(s.Targets)
	for i := 0; i < count; i++ {
		if target.Id == s.Targets[i].Id {
			continue
		}
		if target.Domain == s.Targets[i].Domain && target.Path == s.Targets[i].Path {
			return nil, fmt.Errorf("domain '%s' and path '%s' not existed", target.Domain, target.Path)
		}
	}

	var modifiedTarget *ProxyTarget = nil
	for i := 0; i < count; i++ {
		if target.Id == s.Targets[i].Id {
			s.Targets[i].CopyFrom(target)
			modifiedTarget = s.Targets[i]
			break
		}
	}
	if modifiedTarget == nil {
		return nil, fmt.Errorf("target id '%s' not existed", target.Id)
	}

	return modifiedTarget, nil
}

type ProxyServerAdd struct {
	Name    string `json:"name" required:"true" note:"名称"`
	Disable bool   `json:"disable" note:"已禁用"`
	TLS     bool   `json:"tls" note:"传入是否为TLS连接"`
	IP      string `json:"ip" note:"监听地址，空表示所有IP地址"`
	Port    string `json:"port" required:"true" note:"监听端口"`
}

type ProxyServerDel struct {
	Id string `json:"id" required:"true" note:"标识ID"`
}

type ProxyServerEdit struct {
	ProxyServerDel
	ProxyServerAdd
}

func (s *ProxyServerEdit) CopyTo(target *ProxyServer) {
	if target == nil {
		return
	}

	target.Name = s.Name
	target.Disable = s.Disable
	target.TLS = s.TLS
	target.IP = s.IP
	target.Port = s.Port
}

func (s *ProxyServerEdit) CopyFrom(source *ProxyServer) {
	if source == nil {
		return
	}

	s.Id = source.Id
	s.Name = source.Name
	s.Disable = source.Disable
	s.TLS = source.TLS
	s.IP = source.IP
	s.Port = source.Port
}
