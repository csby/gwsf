package gtype

type NodeId struct {
	Instance    string `json:"instance" note:"实例标识ID"`
	Certificate string `json:"certificate" note:"证书标识ID"`
}

type Node struct {
	ID        NodeId   `json:"id" note:"结点标识ID"`
	Kind      string   `json:"kind" note:"结点类型, 如: client"`
	Name      string   `json:"name" note:"结点名称"`
	IP        string   `json:"ip" note:"结点IP地址"`
	LoginTime DateTime `json:"loginTime" note:"登陆时间"`
}

func (s *Node) CopyFrom(source *Token) {
	if source == nil {
		return
	}

	s.Kind = source.UserAccount
	s.Name = source.UserName
	s.IP = source.LoginIP
	s.LoginTime = DateTime(source.LoginTime)

	id, ok := source.Ext.(NodeId)
	if ok {
		s.ID.Instance = id.Instance
		s.ID.Certificate = id.Certificate
	}
}
