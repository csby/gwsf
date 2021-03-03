package gtype

type Node struct {
	ID        string   `json:"id" note:"节点ID"`
	Kind      string   `json:"kind" note:"节点类型, 如: client"`
	Name      string   `json:"name" note:"节点名称"`
	IP        string   `json:"ip" note:"节点IP地址"`
	LoginTime DateTime `json:"loginTime" note:"登陆时间"`
}

func (s *Node) CopyFrom(source *Token) {
	if source == nil {
		return
	}

	s.ID = source.ID
	s.Kind = source.UserAccount
	s.Name = source.UserName
	s.IP = source.LoginIP
	s.LoginTime = DateTime(source.LoginTime)
}
