package gcloud

import (
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"sync"
	"time"
)

type Node struct {
	Id   string      `json:"id" note:"节点ID"`
	Name string      `json:"name" note:"节点名称"`
	Kind string      `json:"kind" note:"节点类型"`
	Addr NodeAddress `json:"addr" note:"节点地址"`

	RegTime     gtype.DateTime  `json:"regTime" note:"注册时间"`
	RegIp       string          `json:"regIp" note:"注册IP地址"`
	OnlineTime  gtype.DateTime  `json:"onlineTime" note:"最近上线时间"`
	OfflineTime *gtype.DateTime `json:"offlineTime" note:"最近离线时间"`

	DisplayName string `json:"displayName" note:"显示名称"`
	Remark      string `json:"remark" note:"备注信息"`

	Instances []*NodeInstance `json:"instances" note:"在线实例"`
}

func (s *Node) CopyToClient(t *gcfg.HttpsClient) {
	if t == nil {
		return
	}

	t.Id = s.Id
	t.Name = s.Name
	t.Kind = s.Kind
	t.Addr.Province = s.Addr.Address
	t.Addr.Locality = s.Addr.Locality
	t.Addr.Address = s.Addr.Address
	t.RegTime = s.RegTime
	t.RegIp = s.RegIp
	t.OnlineTime = s.OnlineTime
	if t.OfflineTime != nil {
		v := gtype.DateTime(*t.OfflineTime)
		s.OfflineTime = &v
	}
	t.DisplayName = s.DisplayName
	t.Remark = s.Remark
}

func (s *Node) CopyFromClient(src *gcfg.HttpsClient) {
	if src == nil {
		return
	}

	s.Id = src.Id
	s.Name = src.Name
	s.Kind = src.Kind
	s.Addr.Province = src.Addr.Province
	s.Addr.Locality = src.Addr.Locality
	s.Addr.Address = src.Addr.Address
	s.RegTime = src.RegTime
	s.RegIp = src.RegIp
	s.OnlineTime = src.OnlineTime
	if src.OfflineTime != nil {
		t := gtype.DateTime(*src.OfflineTime)
		s.OfflineTime = &t
	}
	s.DisplayName = src.DisplayName
	s.Remark = src.Remark

	s.Instances = make([]*NodeInstance, 0)
}

func (s *Node) CopyFromNode(src *gtype.Node) {
	if src == nil {
		return
	}

	s.Id = src.ID.Certificate
	s.Name = src.Name
	s.Kind = src.Kind
	s.Addr.Province = src.Province
	s.Addr.Locality = src.Locality
	s.Addr.Address = src.Address
	s.RegTime = src.LoginTime
	s.RegIp = src.IP
	s.OnlineTime = src.LoginTime
	s.DisplayName = src.Name

	s.Instances = make([]*NodeInstance, 0)
}

type NodeAddress struct {
	Province string `json:"province" note:"省份"`
	Locality string `json:"locality" note:"地区"`
	Address  string `json:"address" note:"地址"`
}

type NodeInstance struct {
	Node        string          `json:"node" note:"节点ID"`
	Id          string          `json:"id" note:"实例ID"`
	Ip          string          `json:"ip" note:"IP地址"`
	Version     string          `json:"version" note:"版本号"`
	Time        gtype.DateTime  `json:"time" note:"上线时间"`
	CrtNotAfter *gtype.DateTime `json:"crtNotAfter" note:"证书到期时间"`
}

type NodeModify struct {
	Id          string `json:"id" required:"true" note:"节点ID"`
	DisplayName string `json:"displayName" note:"显示名称"`
	Remark      string `json:"remark" note:"备注信息"`
}

type NodeDelete struct {
	Id string `json:"id" required:"true" note:"节点ID"`
}

type NodeCollection struct {
	sync.RWMutex

	items []*Node

	cfg *gcfg.Config
	opt gtype.SocketChannelCollection
}

func (s *NodeCollection) Items() []*Node {
	return s.items
}

func (s *NodeCollection) InitFromCfg(cfg *gcfg.Config) {
	s.Lock()
	defer s.Unlock()

	if s.items == nil {
		s.items = make([]*Node, 0)
	}

	if cfg == nil {
		return
	}
	items := cfg.Cloud.Clients
	c := len(items)
	for i := 0; i < c; i++ {
		item := items[i]
		if item == nil {
			continue
		}

		node := &Node{}
		node.CopyFromClient(item)
		s.items = append(s.items, node)
	}
}

func (s *NodeCollection) NodeOnline(v *gtype.Node) {
	if v == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	nodeId := v.ID.Certificate
	instId := v.ID.Instance
	var node *Node = nil
	count := len(s.items)
	for index := 0; index < count; index++ {
		item := s.items[index]
		if item == nil {
			continue
		}
		if item.Id == nodeId {
			node = item
			break
		}
	}

	if node == nil {
		inst := &NodeInstance{
			Node:        nodeId,
			Id:          instId,
			Ip:          v.IP,
			Time:        v.LoginTime,
			Version:     v.Version,
			CrtNotAfter: v.CrtNotAfter,
		}

		node = &Node{}
		node.CopyFromNode(v)
		node.Instances = append(node.Instances, inst)
		s.items = append(s.items, node)

		s.SaveToCfg()
		s.writeOptSocketMessage(gtype.WSNodeRegister, node)
	} else {
		inst := &NodeInstance{
			Node:        nodeId,
			Id:          instId,
			Ip:          v.IP,
			Time:        v.LoginTime,
			Version:     v.Version,
			CrtNotAfter: v.CrtNotAfter,
		}
		if len(node.Instances) < 1 {
			node.OnlineTime = v.LoginTime
			s.SaveToCfg()
		}
		node.Instances = append(node.Instances, inst)
		s.writeOptSocketMessage(gtype.WSNodeInstanceOnline, inst)
	}
}

func (s *NodeCollection) NodeOffline(v *gtype.Node) {
	if v == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	nodeId := v.ID.Certificate
	instId := v.ID.Instance
	var node *Node = nil
	count := len(s.items)
	for index := 0; index < count; index++ {
		item := s.items[index]
		if item == nil {
			continue
		}
		if item.Id == nodeId {
			node = item
			break
		}
	}

	if node == nil {
		return
	}

	count = len(node.Instances)
	if count < 1 {
		return
	}

	var inst *NodeInstance = nil
	instances := make([]*NodeInstance, 0)
	for index := 0; index < count; index++ {
		item := node.Instances[index]
		if item == nil {
			continue
		}
		if item.Id == instId {
			inst = item
		} else {
			instances = append(instances, item)
		}
	}

	if inst == nil {
		return
	}
	node.Instances = instances

	now := gtype.DateTime(time.Now())
	inst.Time = now
	if len(node.Instances) < 1 {
		node.OfflineTime = &now
		s.SaveToCfg()
	}
	s.writeOptSocketMessage(gtype.WSNodeInstanceOffline, inst)
}

func (s *NodeCollection) NodeModify(v *NodeModify) error {
	if v == nil {
		return fmt.Errorf("v is nil")
	}

	s.Lock()
	defer s.Unlock()

	nodeId := v.Id
	var node *Node = nil
	count := len(s.items)
	for index := 0; index < count; index++ {
		item := s.items[index]
		if item == nil {
			continue
		}
		if item.Id == nodeId {
			node = item
			break
		}
	}
	if node == nil {
		return fmt.Errorf("节点(id=%s)不存在", nodeId)
	}

	affectedCount := 0
	if node.DisplayName != v.DisplayName {
		node.DisplayName = v.DisplayName
		affectedCount++
	}
	if node.Remark != v.Remark {
		node.Remark = v.Remark
		affectedCount++
	}
	if affectedCount < 1 {
		return nil
	}

	s.SaveToCfg()
	s.writeOptSocketMessage(gtype.WSNodeModify, v)

	return nil
}

func (s *NodeCollection) NodeDelete(id string) gtype.Error {
	s.Lock()
	defer s.Unlock()

	nodeId := id
	var node *Node = nil
	nodes := make([]*Node, 0)
	count := len(s.items)
	for index := 0; index < count; index++ {
		item := s.items[index]
		if item == nil {
			continue
		}
		if item.Id == nodeId {
			node = item
		} else {
			nodes = append(nodes, item)
		}
	}
	if node == nil {
		return gtype.ErrInput.SetDetail(fmt.Sprintf("节点(id=%s)不存在", nodeId))
	}
	if len(node.Instances) > 0 {
		return gtype.ErrNotSupport.SetDetail("不允许删除在线节点")
	}

	s.items = nodes
	s.SaveToCfg()
	s.writeOptSocketMessage(gtype.WSNodeRevoke, &NodeDelete{Id: nodeId})

	return nil
}

func (s *NodeCollection) SaveToCfg() error {
	if s.cfg == nil {
		return fmt.Errorf("cfg is nil")
	}

	if s.cfg.Load == nil {
		return fmt.Errorf("load not config")
	}
	if s.cfg.Save == nil {
		return fmt.Errorf("save not config")
	}
	cfg, err := s.cfg.Load()
	if err != nil {
		return err
	}

	clients := make([]*gcfg.HttpsClient, 0)
	items := s.items
	c := len(items)
	for i := 0; i < c; i++ {
		item := items[i]
		if item == nil {
			continue
		}
		client := &gcfg.HttpsClient{}
		item.CopyToClient(client)
		clients = append(clients, client)
	}
	cfg.Cloud.Clients = clients
	err = s.cfg.Save(cfg)
	if err != nil {
		return fmt.Errorf("save config fail: %s", err.Error())
	}
	s.cfg.Cloud.Clients = clients

	return nil
}

func (s *NodeCollection) writeOptSocketMessage(id int, data interface{}) bool {
	if s.opt == nil {
		return false
	}

	msg := &gtype.SocketMessage{
		ID:   id,
		Data: data,
	}

	s.opt.Write(msg, nil)

	return true
}
