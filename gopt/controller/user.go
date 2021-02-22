package controller

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"time"
)

type User struct {
	controller
}

func NewUser(log gtype.Log, cfg *gcfg.Config, db gtype.TokenDatabase, chs gtype.SocketChannelCollection) *User {
	instance := &User{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.dbToken = db
	instance.wsChannels = chs

	return instance
}

func (s *User) GetLoginAccount(ctx gtype.Context, ps gtype.Params) {
	token := s.getToken(ctx.Token())
	if token == nil {
		ctx.Error(gtype.ErrInternal, "凭证无效")
		return
	}
	ctx.Success(&gtype.LoginAccount{
		Account:   token.UserAccount,
		Name:      token.UserName,
		LoginTime: gtype.DateTime(token.LoginTime),
	})
}

func (s *User) GetLoginAccountDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "用户管理")
	function := catalog.AddFunction(method, uri, "获取登录账号")
	function.SetNote("获取当前登录账号基本信息")
	function.SetOutputDataExample(&gtype.LoginAccount{
		Account:   "admin",
		Name:      "管理员",
		LoginTime: gtype.DateTime(time.Now()),
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *User) GetOnlineUsers(ctx gtype.Context, ps gtype.Params) {
	if s.wsChannels == nil {
		ctx.Error(gtype.ErrInternal, "websocket channels is nil")
		return
	}
	ctx.Success(s.wsChannels.OnlineUsers())
}

func (s *User) GetOnlineUsersDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "用户管理")
	function := catalog.AddFunction(method, uri, "获取在线用户")
	function.SetNote("获取当前所有在线用户")
	function.SetOutputDataExample([]gtype.OnlineUser{
		{
			UserAccount:   "admin",
			UserName:      "管理员",
			LoginIP:       "192.168.1.8",
			LoginTime:     gtype.DateTime(time.Now()),
			LoginDuration: "7秒",
		},
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}
