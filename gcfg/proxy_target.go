package gcfg

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
	"sync"
)

type ProxySpare struct {
	sync.RWMutex

	AddrId    string `json:"addrId" note:"地址标识"`
	Alive     bool   `json:"alive" note:"在线状态"`
	ConnCount int64  `json:"connCount" note:"连接数量"`
	IP        string `json:"ip" note:"目标地址"`
	Port      string `json:"port" note:"目标端口"`

	sourceId string
	targetId string
}

func (s *ProxySpare) SetSourceId(v string) {
	s.sourceId = v
}

func (s *ProxySpare) SetTargetId(v string) {
	s.targetId = v
}

func (s *ProxySpare) SourceId() string {
	return s.sourceId
}

func (s *ProxySpare) TargetId() string {
	return s.targetId
}

func (s *ProxySpare) IsAlive() bool {
	return s.Alive
}

func (s *ProxySpare) Count() int64 {
	return s.ConnCount
}

func (s *ProxySpare) SetAlive(v bool) {
	s.Alive = v
}

func (s *ProxySpare) IncreaseCount() int64 {
	s.Lock()
	defer s.Unlock()

	s.ConnCount += 1

	return s.ConnCount
}

func (s *ProxySpare) DecreaseCount() int64 {
	s.Lock()
	defer s.Unlock()

	s.ConnCount -= 1

	return s.ConnCount
}

type ProxyTarget struct {
	sync.RWMutex

	Id     string `json:"id" note:"标识ID"`
	Domain string `json:"domain" note:"域名"`
	Path   string `json:"path" note:"路径，仅http有效"`

	AddrId    string        `json:"addrId" note:"地址标识"`
	Alive     bool          `json:"alive" note:"在线状态"`
	ConnCount int64         `json:"connCount" note:"连接数量"`
	IP        string        `json:"ip" note:"目标地址"`
	Port      string        `json:"port" note:"目标端口"`
	Version   int           `json:"version" note:"版本号，0或1，0-不添加头部；1-添加代理头部（PROXY family srcIP srcPort targetIP targetPort）"`
	Disable   bool          `json:"disable" note:"已禁用"`
	Spares    []*ProxySpare `json:"spares" note:"备用目标"`

	sourceId string
}

func (s *ProxyTarget) SetSourceId(v string) {
	s.sourceId = v
}

func (s *ProxyTarget) SourceId() string {
	return s.sourceId
}

func (s *ProxyTarget) TargetId() string {
	return s.Id
}

func (s *ProxyTarget) InitAddrId(sourceId string) {
	val := fmt.Sprintf("%s-%s-%s:%s",
		sourceId, s.Id, s.IP, s.Port)
	s.AddrId = gtype.ToMd5(val)

	c := len(s.Spares)
	for i := 0; i < c; i++ {
		item := s.Spares[i]
		if item == nil {
			continue
		}

		val = fmt.Sprintf("%s-%s-%s:%s",
			sourceId, s.Id, item.IP, item.Port)
		item.AddrId = gtype.ToMd5(val)
	}
}

func (s *ProxyTarget) CopyFrom(source *ProxyTarget) {
	if source == nil {
		return
	}

	s.Domain = source.Domain
	s.Path = source.Path
	s.IP = source.IP
	s.Port = source.Port
	s.Version = source.Version
	s.Disable = source.Disable
	s.Spares = make([]*ProxySpare, 0)
	for i := 0; i < len(source.Spares); i++ {
		item := source.Spares[i]
		if item != nil {
			s.Spares = append(s.Spares, &ProxySpare{
				IP:   item.IP,
				Port: item.Port,
			})
		}
	}
}

func (s *ProxyTarget) SpareTargets() []string {
	targets := make([]string, 0)

	c := len(s.Spares)
	for i := 0; i < c; i++ {
		spare := s.Spares[i]
		if spare == nil {
			continue
		}

		targets = append(targets, fmt.Sprintf("%s:%s", spare.IP, spare.Port))
	}

	return targets
}

func (s *ProxyTarget) IsAlive() bool {
	return s.Alive
}

func (s *ProxyTarget) Count() int64 {
	return s.ConnCount
}

func (s *ProxyTarget) SetAlive(v bool) {
	s.Alive = v
}

func (s *ProxyTarget) IncreaseCount() int64 {
	s.Lock()
	defer s.Unlock()

	s.ConnCount += 1

	return s.ConnCount
}

func (s *ProxyTarget) DecreaseCount() int64 {
	s.Lock()
	defer s.Unlock()

	s.ConnCount -= 1
	if s.ConnCount < 0 {
		s.ConnCount = 0
	}

	return s.ConnCount
}

type ProxyTargetEdit struct {
	ServerId string      `json:"serverId" required:"true" note:"服务器标识ID"`
	Target   ProxyTarget `json:"target" note:"目标地址"`
}

type ProxyTargetDel struct {
	ServerId string `json:"serverId" required:"true" note:"服务器标识ID"`
	TargetId string `json:"targetId" required:"true" note:"目标地址标识ID"`
}
