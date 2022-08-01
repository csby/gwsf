package gcfg

type SiteOptLdap struct {
	Enable bool   `json:"enable" note:"是否启用"`
	Host   string `json:"host" note:"主机地址"`
	Port   int    `json:"port" note:"端口号，如389"`
	Base   string `json:"base" note:"位置，如‘dc=example,dc=com’"`
}

func (s *SiteOptLdap) CopyTo(target *SiteOptLdap) int {
	if target == nil {
		return 0
	}

	count := 0
	if target.Enable != s.Enable {
		target.Enable = s.Enable
		count++
	}
	if target.Host != s.Host {
		target.Host = s.Host
		count++
	}
	if target.Port != s.Port {
		target.Port = s.Port
		count++
	}
	if target.Base != s.Base {
		target.Base = s.Base
		count++
	}

	return count
}
