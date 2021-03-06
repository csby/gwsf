package gtype

import (
	"container/list"
	"encoding/json"
	"sync"
)

const (
	WSOptUserLogin  = 101 // 用户登陆
	WSOptUserLogout = 102 // 用户注销

	WSSiteUpload         = 110 // 上传并发布应用网站
	WSRootSiteUploadFile = 111 // 根站点-上传文件
	WSRootSiteDeleteFile = 112 // 根站点-删除文件

	WSNodeOnline             = 121 // 结点上线
	WSNodeOffline            = 122 // 结点下线
	WSNodeForwardTcpStart    = 123 // 结点转发TCP开始
	WSNodeForwardTcpEnd      = 124 // 结点转发TCP结束
	WSNodeForwardUdpRequest  = 125 // 结点转发UDP请求
	WSNodeForwardUdpResponse = 126 // 结点转发UDP响应
)

type SocketMessage struct {
	ID   int         `json:"id" note:"消息标识"`
	Data interface{} `json:"data" note:"消息内容, 结构随id而定"`
}

func (s *SocketMessage) GetData(v interface{}) error {
	data, err := json.Marshal(s.Data)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

type SocketChannel interface {
	Token() *Token
	Container() SocketChannelCollection
	Write(message *SocketMessage)
	Read() <-chan *SocketMessage

	getElement() *list.Element
	close()
}

type innerSocketChannel struct {
	channel   chan *SocketMessage
	element   *list.Element
	container *innerSocketChannelCollection
	token     *Token
}

func (s *innerSocketChannel) Token() *Token {
	return s.token
}

func (s *innerSocketChannel) Container() SocketChannelCollection {
	return s.container
}

func (s *innerSocketChannel) Write(message *SocketMessage) {
	select {
	case s.channel <- message:
	default:
	}
}

func (s *innerSocketChannel) Read() <-chan *SocketMessage {
	return s.channel
}

func (s *innerSocketChannel) getElement() *list.Element {
	return s.element
}

func (s *innerSocketChannel) close() {
	close(s.channel)
}

type SocketChannelCollection interface {
	OnlineUsers() []*OnlineUser
	OnlineNodes() []*Node
	OnlineNode(id NodeId) *Node
	SetListener(newChannel, removeChannel func(channel SocketChannel))
	NewChannel(token *Token) SocketChannel
	Remove(channel SocketChannel)
	Write(message *SocketMessage, token *Token)
	WriteMessage(message *SocketMessage, tokenId string) bool
	AddReader(reader func(message *SocketMessage, channel SocketChannel))
	Read(message *SocketMessage, channel SocketChannel)
	AddFilter(filter func(message *SocketMessage, channel SocketChannel, token *Token) bool)
}

func NewSocketChannelCollection() SocketChannelCollection {
	instance := &innerSocketChannelCollection{}
	instance.channels = list.New()
	instance.readers = make([]func(message *SocketMessage, channel SocketChannel), 0)
	instance.filters = make([]func(message *SocketMessage, channel SocketChannel, token *Token) bool, 0)

	return instance
}

type innerSocketChannelCollection struct {
	sync.RWMutex

	channels       *list.List
	readers        []func(message *SocketMessage, channel SocketChannel)
	filters        []func(message *SocketMessage, channel SocketChannel, token *Token) bool
	newListener    func(channel SocketChannel)
	removeListener func(channel SocketChannel)
}

func (s *innerSocketChannelCollection) OnlineUsers() []*OnlineUser {
	s.Lock()
	defer s.Unlock()

	tokens := make(map[string]int)
	users := make([]*OnlineUser, 0)

	for e := s.channels.Front(); e != nil; {
		ev, ok := e.Value.(SocketChannel)
		if !ok {
			break
		}

		token := ev.Token()
		if token != nil {
			_, ok := tokens[token.ID]
			if !ok {
				tokens[token.ID] = 0
				user := &OnlineUser{}
				user.CopyFrom(token)
				user.LoginDuration = user.LoginTime.Duration()

				users = append(users, user)
			}
		}

		e = e.Next()
	}

	return users
}

func (s *innerSocketChannelCollection) OnlineNodes() []*Node {
	s.Lock()
	defer s.Unlock()

	nodes := make([]*Node, 0)
	for e := s.channels.Front(); e != nil; {
		ev, ok := e.Value.(SocketChannel)
		if !ok {
			break
		}

		token := ev.Token()
		if token != nil {
			node := &Node{}
			node.CopyFrom(token)
			nodes = append(nodes, node)
		}

		e = e.Next()
	}

	return nodes
}

func (s *innerSocketChannelCollection) OnlineNode(id NodeId) *Node {
	s.Lock()
	defer s.Unlock()

	for e := s.channels.Front(); e != nil; {
		ev, ok := e.Value.(SocketChannel)
		if !ok {
			break
		}

		token := ev.Token()
		if token != nil {
			if len(id.Instance) > 0 {
				if id.Instance == token.ID {
					node := &Node{}
					node.CopyFrom(token)
					return node
				}
			} else if len(id.Certificate) > 0 {
				nodeId, ok := token.Ext.(NodeId)
				if ok {
					if id.Certificate == nodeId.Certificate {
						node := &Node{}
						node.CopyFrom(token)
						return node
					}
				}
			}

		}

		e = e.Next()
	}

	return nil
}

func (s *innerSocketChannelCollection) SetListener(newChannel, removeChannel func(channel SocketChannel)) {
	s.newListener = newChannel
	s.removeListener = removeChannel
}

func (s *innerSocketChannelCollection) NewChannel(token *Token) SocketChannel {
	s.Lock()
	defer s.Unlock()

	instance := &innerSocketChannel{container: s}
	instance.channel = make(chan *SocketMessage, 1024)
	instance.element = s.channels.PushBack(instance)
	instance.token = token
	if token != nil {
		token.Usage++
	}

	if s.newListener != nil {
		s.newListener(instance)
	}

	return instance
}

func (s *innerSocketChannelCollection) Remove(channel SocketChannel) {
	if channel == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	token := channel.Token()
	if token != nil {
		token.Usage--
	}

	if s.removeListener != nil {
		s.removeListener(channel)
	}

	s.channels.Remove(channel.getElement())
	channel.close()
}

func (s *innerSocketChannelCollection) Write(message *SocketMessage, token *Token) {
	s.Lock()
	defer s.Unlock()

	for e := s.channels.Front(); e != nil; {
		ev, ok := e.Value.(SocketChannel)
		if !ok {
			return
		}

		if !s.filter(message, ev, token) {
			ev.Write(message)
		}

		e = e.Next()
	}
}

func (s *innerSocketChannelCollection) WriteMessage(message *SocketMessage, tokenId string) bool {
	s.Lock()
	defer s.Unlock()

	for e := s.channels.Front(); e != nil; {
		ev, ok := e.Value.(SocketChannel)
		if !ok {
			return false
		}

		t := ev.Token()
		if t != nil {
			if t.ID == tokenId {
				ev.Write(message)
				return true
			}
		}

		e = e.Next()
	}

	return false
}

func (s *innerSocketChannelCollection) AddReader(reader func(message *SocketMessage, channel SocketChannel)) {
	if reader == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.readers = append(s.readers, reader)
}

func (s *innerSocketChannelCollection) Read(message *SocketMessage, channel SocketChannel) {
	count := len(s.readers)
	for i := 0; i < count; i++ {
		reader := s.readers[0]
		if reader == nil {
			continue
		}

		go func(read func(message *SocketMessage, channel SocketChannel), msg *SocketMessage, ch SocketChannel) {
			defer func() {
				if err := recover(); err != nil {
				}
			}()

			read(msg, ch)
		}(reader, message, channel)
	}
}

func (s *innerSocketChannelCollection) AddFilter(filter func(message *SocketMessage, channel SocketChannel, token *Token) bool) {
	if filter == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.filters = append(s.filters, filter)
}

func (s *innerSocketChannelCollection) filter(message *SocketMessage, channel SocketChannel, token *Token) bool {
	count := len(s.filters)
	for i := 0; i < count; i++ {
		filter := s.filters[0]
		if filter == nil {
			continue
		}

		if filter(message, channel, token) {
			return true
		}
	}

	return false
}
