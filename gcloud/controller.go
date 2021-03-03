package gcloud

import (
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"net/http"
)

type Controller struct {
	gtype.Base

	cfg *gcfg.Config
	chs *Channels

	wsGrader websocket.Upgrader
}

func NewController(log gtype.Log, cfg *gcfg.Config, chs *Channels) *Controller {
	instance := &Controller{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.chs = chs
	instance.wsGrader = websocket.Upgrader{CheckOrigin: instance.checkOrigin}

	return instance
}

func (s *Controller) preHandle(ctx gtype.Context, ps gtype.Params) {
	schema := ctx.Schema()
	if schema != "https" {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("protocol '%s' not support, 'https' is required", schema))
		ctx.SetHandled(true)
		return
	}

	crt := ctx.Certificate().Client
	if crt == nil {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("certificate is required"))
		ctx.SetHandled(true)
		return
	}
	ou := crt.OrganizationalUnit()
	if len(ou) < 1 {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("organization unit of certificate is empty"))
		ctx.SetHandled(true)
		return
	}
}

func (s *Controller) createCatalog(doc gtype.Doc, names ...string) gtype.Catalog {
	root := doc.AddCatalog("云平台接口")

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

func (s *Controller) checkOrigin(r *http.Request) bool {
	if r != nil {
	}

	return true
}

func (s *Controller) writeOptSocketMessage(id int, data interface{}) bool {
	if s.chs == nil {
		return false
	}
	if s.chs.opt == nil {
		return false
	}

	msg := &gtype.SocketMessage{
		ID:   id,
		Data: data,
	}

	s.chs.opt.Write(msg, nil)

	return true
}
