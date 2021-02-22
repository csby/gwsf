package controller

import (
	"crypto/rand"
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"io"
)

type controller struct {
	gtype.Base

	cfg *gcfg.Config

	dbToken    gtype.TokenDatabase
	wsChannels gtype.SocketChannelCollection
}

func (s *controller) createCatalog(doc gtype.Doc, names ...string) gtype.Catalog {
	root := doc.AddCatalog("管理平台接口")

	count := len(names)
	if count < 1 {
		return root
	}

	child := root
	for i := 0; i < count; i++ {
		name := names[i]
		child = child.AddChild(name)
	}

	return child
}

func (s *controller) getToken(key string) *gtype.Token {
	if len(key) < 1 {
		return nil
	}

	if s.dbToken == nil {
		return nil
	}

	value, ok := s.dbToken.Get(key, false)
	if !ok {
		return nil
	}

	token, ok := value.(*gtype.Token)
	if !ok {
		return nil
	}

	return token
}

func (s *controller) writeWebSocketMessage(token string, id int, data interface{}) bool {
	if s.wsChannels == nil {
		return false
	}

	msg := &gtype.SocketMessage{
		ID:   id,
		Data: data,
	}

	s.wsChannels.Write(msg, s.getToken(token))

	return true
}

func (s *controller) newGuid() string {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return ""
	}

	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40

	return fmt.Sprintf("%x%x%x%x%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
