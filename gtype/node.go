package gtype

type NodeId struct {
	Instance    string `json:"instance" note:"实例标识ID"`
	Certificate string `json:"certificate" note:"证书标识ID"`
}

type Node struct {
	ID        NodeId   `json:"id" note:"节点标识ID"`
	Kind      string   `json:"kind" note:"节点类型, 如: client"`
	Name      string   `json:"name" note:"节点名称"`
	IP        string   `json:"ip" note:"节点IP地址"`
	Version   string   `json:"version" note:"版本号"`
	LoginTime DateTime `json:"loginTime" note:"登陆时间"`

	Province    string    `json:"province" note:"省份, 对于证书S"`
	Locality    string    `json:"locality" note:"地区, 对于证书L"`
	Address     string    `json:"address" note:"地址, 对于证书STREET""`
	CrtNotAfter *DateTime `json:"crtNotAfter" note:"证书到期时间"`
}

func (s *Node) CopyFrom(source *Token) {
	if source == nil {
		return
	}

	s.Kind = source.UserAccount
	s.Name = source.UserName
	s.IP = source.LoginIP
	s.Version = source.Version
	s.LoginTime = DateTime(source.LoginTime)

	id, ok := source.Ext.(NodeId)
	if ok {
		s.ID.Instance = id.Instance
		s.ID.Certificate = id.Certificate
	}
}

type NodeInfo struct {
	NodeId

	Name         string `json:"name" note:"节点名称"`
	Enabled      bool   `json:"enabled" note:"是否启用"`
	Online       bool   `json:"online" note:"true-在线; false-离线"`
	CloudAddress string `json:"cloudAddress" note:"云服务地址"`
}
