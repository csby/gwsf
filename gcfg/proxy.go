package gcfg

import "fmt"

type Proxy struct {
	Enabled bool           `json:"enabled" note:"是否启用(是否显示页面)"`
	Disable bool           `json:"disable" note:"已禁用"`
	Servers []*ProxyServer `json:"servers" note:"服务器"`
}

func (s *Proxy) initId() {
	count := len(s.Servers)
	for i := 0; i < count; i++ {
		item := s.Servers[i]
		if item == nil {
			continue
		}
		item.initId()
	}
}

func (s *Proxy) CopyTo(target *Proxy) {
	if target == nil {
		return
	}

	target.Disable = s.Disable
	target.Servers = s.Servers
}

func (s *Proxy) AddServer(server *ProxyServer) error {
	if server == nil {
		return fmt.Errorf("server is nil")
	}

	id := server.UniqueId()
	count := len(s.Servers)
	for i := 0; i < count; i++ {
		if id == s.Servers[i].UniqueId() {
			return fmt.Errorf("server '%s' has been existed", id)
		}
	}

	s.Servers = append(s.Servers, server)

	return nil
}

func (s *Proxy) DeleteServer(server *ProxyServer) error {
	if server == nil {
		return fmt.Errorf("server is nil")
	}

	id := server.Id
	servers := make([]*ProxyServer, 0)
	count := len(s.Servers)
	deletedCount := 0
	for i := 0; i < count; i++ {
		if id == s.Servers[i].Id {
			deletedCount++
			continue
		}
		servers = append(servers, s.Servers[i])
	}
	if deletedCount <= 0 {
		return fmt.Errorf("server id '%s' not existed", id)
	}

	s.Servers = servers

	return nil
}

func (s *Proxy) ModifyServer(server *ProxyServerEdit) error {
	if server == nil {
		return fmt.Errorf("server is nil")
	}

	var item *ProxyServer = nil
	id := server.Id
	uid := fmt.Sprintf("%s:%s", server.IP, server.Port)
	count := len(s.Servers)
	for i := 0; i < count; i++ {
		srv := s.Servers[i]
		if id == srv.Id {
			item = srv
			continue
		}

		srvUid := srv.UniqueId()
		if uid == srvUid {
			return fmt.Errorf("server '%s' has been existed", uid)
		}
	}

	if item == nil {
		return fmt.Errorf("server id '%s' not existed", id)
	}

	server.CopyTo(item)

	return nil
}

func (s *Proxy) GetServer(id string) *ProxyServer {
	count := len(s.Servers)
	for i := 0; i < count; i++ {
		server := s.Servers[i]
		if server == nil {
			continue
		}
		if id == server.Id {
			return server
		}
	}

	return nil
}
