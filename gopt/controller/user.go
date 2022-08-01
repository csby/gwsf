package controller

import (
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"strings"
	"time"
)

const (
	adminAccount = "admin"
)

const (
	roleAdmin = 1
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
	role := 0
	if strings.ToLower(token.UserAccount) == "admin" {
		role = roleAdmin
	}
	ctx.Success(&gtype.LoginAccount{
		Account:   token.UserAccount,
		Name:      token.UserName,
		LoginTime: gtype.DateTime(token.LoginTime),
		Role:      role,
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

func (s *User) GetList(ctx gtype.Context, ps gtype.Params) {
	results := make([]gtype.AccountInfo, 0)
	if s.cfg != nil {
		users := s.cfg.Site.Opt.Users
		c := len(users)
		for i := 0; i < c; i++ {
			user := users[i]
			result := gtype.AccountInfo{
				Account: user.Account,
				Name:    user.Name,
			}
			if strings.ToLower(user.Account) == "admin" {
				result.BuiltIn = true
			}
			results = append(results, result)
		}
	}

	ctx.Success(results)
}

func (s *User) GetListDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "用户管理")
	function := catalog.AddFunction(method, uri, "获取本地用户列表")
	function.SetNote("获取所有用户")
	function.SetOutputDataExample([]gtype.AccountInfo{
		{
			Account: "admin",
			Name:    "管理员",
			BuiltIn: true,
		},
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *User) Create(ctx gtype.Context, ps gtype.Params) {
	if s.cfg == nil {
		ctx.Error(gtype.ErrInternal, "cfg is nil")
		return
	}
	if s.cfg.Load == nil {
		ctx.Error(gtype.ErrInternal, "load not config")
		return
	}
	if s.cfg.Save == nil {
		ctx.Error(gtype.ErrInternal, "save not config")
		return
	}

	token := s.getToken(ctx.Token())
	if token == nil {
		ctx.Error(gtype.ErrInternal, "凭证无效")
		return
	}
	if strings.ToLower(token.UserAccount) != adminAccount {
		ctx.Error(gtype.ErrNoPermission, fmt.Sprintf("只有内置管理员帐号(%s)才能新建用户", adminAccount))
		return
	}

	argument := &gtype.AccountCreate{}
	err := ctx.GetJson(&argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	account := strings.TrimSpace(argument.Account)
	if len(account) < 1 {
		ctx.Error(gtype.ErrInput, "帐号为空")
		return
	}
	site := &s.cfg.Site.Opt
	if site.GetUser(account) != nil {
		ctx.Error(gtype.ErrExist, fmt.Sprintf("帐号(%s)已存在", account))
		return
	}

	users := site.Users
	if users == nil {
		users = make([]*gcfg.SiteOptUser, 0)
	}
	users = append(users, &gcfg.SiteOptUser{
		Account:  account,
		Password: strings.TrimSpace(argument.Password),
		Name:     strings.TrimSpace(argument.Name),
	})
	cfg, err := s.cfg.Load()
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Errorf("load config fail: %s", err.Error()))
		return
	}
	cfg.Site.Opt.Users = users
	err = s.cfg.Save(cfg)
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Errorf("save config fail: %s", err.Error()))
		return
	}

	site.Users = users
	ctx.Success(&gtype.AccountInfo{
		Account: account,
		Name:    strings.TrimSpace(argument.Name),
	})
}

func (s *User) CreateDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "用户管理")
	function := catalog.AddFunction(method, uri, "新建本地用户")
	function.SetNote("创建新的系统用户")
	function.SetInputJsonExample(&gtype.AccountCreate{
		Account: "zs",
		Name:    "张三",
	})
	function.SetOutputDataExample(&gtype.AccountInfo{
		Account: "zs",
		Name:    "张三",
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrNoPermission)
	function.AddOutputError(gtype.ErrExist)
}

func (s *User) Modify(ctx gtype.Context, ps gtype.Params) {
	if s.cfg == nil {
		ctx.Error(gtype.ErrInternal, "cfg is nil")
		return
	}
	if s.cfg.Load == nil {
		ctx.Error(gtype.ErrInternal, "load not config")
		return
	}
	if s.cfg.Save == nil {
		ctx.Error(gtype.ErrInternal, "save not config")
		return
	}

	token := s.getToken(ctx.Token())
	if token == nil {
		ctx.Error(gtype.ErrInternal, "凭证无效")
		return
	}

	argument := &gtype.AccountEdit{}
	err := ctx.GetJson(&argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	account := strings.TrimSpace(argument.Account)
	if len(account) < 1 {
		ctx.Error(gtype.ErrInput, "帐号为空")
		return
	}
	if strings.ToLower(token.UserAccount) != strings.ToLower(account) {
		if strings.ToLower(token.UserAccount) != adminAccount {
			ctx.Error(gtype.ErrNoPermission, fmt.Sprintf("只有内置管理员帐号(%s)才能修改其他用户的基本信息", adminAccount))
			return
		}
	}

	site := &s.cfg.Site.Opt
	user := site.GetUser(account)
	if user == nil {
		ctx.Error(gtype.ErrNotExist, fmt.Sprintf("帐号(%s)不存在", account))
		return
	}
	name := strings.TrimSpace(argument.Name)

	cfg, err := s.cfg.Load()
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Errorf("load config fail: %s", err.Error()))
		return
	}
	cfgUser := cfg.Site.Opt.GetUser(account)
	if cfgUser == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("帐号(%s)不存在", account))
		return
	}
	cfgUser.Name = name
	err = s.cfg.Save(cfg)
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Errorf("save config fail: %s", err.Error()))
		return
	}

	user.Name = name
	ctx.Success(nil)
}

func (s *User) ModifyDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "用户管理")
	function := catalog.AddFunction(method, uri, "修改本地用户信息")
	function.SetNote("修改本地用户姓名等基本信息")
	function.SetInputJsonExample(&gtype.AccountEdit{
		Account: "zs",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrNoPermission)
	function.AddOutputError(gtype.ErrExist)
}

func (s *User) Delete(ctx gtype.Context, ps gtype.Params) {
	if s.cfg == nil {
		ctx.Error(gtype.ErrInternal, "cfg is nil")
		return
	}
	if s.cfg.Load == nil {
		ctx.Error(gtype.ErrInternal, "load not config")
		return
	}
	if s.cfg.Save == nil {
		ctx.Error(gtype.ErrInternal, "save not config")
		return
	}

	token := s.getToken(ctx.Token())
	if token == nil {
		ctx.Error(gtype.ErrInternal, "凭证无效")
		return
	}
	if strings.ToLower(token.UserAccount) != adminAccount {
		ctx.Error(gtype.ErrNoPermission, fmt.Sprintf("只有内置管理员帐号(%s)才能删除其他用户", adminAccount))
		return
	}

	argument := &gtype.AccountDelete{}
	err := ctx.GetJson(&argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	account := strings.TrimSpace(argument.Account)
	if len(account) < 1 {
		ctx.Error(gtype.ErrInput, "帐号为空")
		return
	}
	if account == adminAccount {
		ctx.Error(gtype.ErrNotSupport, fmt.Sprintf("不能删除内置管理员帐号(%s)", adminAccount))
		return
	}

	site := &s.cfg.Site.Opt
	user := site.GetUser(account)
	if user == nil {
		ctx.Error(gtype.ErrNotExist, fmt.Sprintf("帐号(%s)不存在", account))
		return
	}

	cfg, err := s.cfg.Load()
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Errorf("load config fail: %s", err.Error()))
		return
	}
	cfgSite := &cfg.Site.Opt
	if cfgSite.RemoveUser(account) < 1 {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("帐号(%s)不存在", account))
		return
	}
	err = s.cfg.Save(cfg)
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Errorf("save config fail: %s", err.Error()))
		return
	}

	site.RemoveUser(account)
	ctx.Success(argument.Account)
}

func (s *User) DeleteDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "用户管理")
	function := catalog.AddFunction(method, uri, "删除本地用户")
	function.SetNote("删除本地用户, 内置管理才能操作")
	function.SetInputJsonExample(&gtype.AccountDelete{
		Account: "zs",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrNoPermission)
	function.AddOutputError(gtype.ErrExist)
}

func (s *User) ResetPassword(ctx gtype.Context, ps gtype.Params) {
	if s.cfg == nil {
		ctx.Error(gtype.ErrInternal, "cfg is nil")
		return
	}
	if s.cfg.Load == nil {
		ctx.Error(gtype.ErrInternal, "load not config")
		return
	}
	if s.cfg.Save == nil {
		ctx.Error(gtype.ErrInternal, "save not config")
		return
	}

	token := s.getToken(ctx.Token())
	if token == nil {
		ctx.Error(gtype.ErrInternal, "凭证无效")
		return
	}
	if strings.ToLower(token.UserAccount) != adminAccount {
		ctx.Error(gtype.ErrNoPermission, fmt.Sprintf("只有内置管理员帐号(%s)才能重置用户密码", adminAccount))
		return
	}

	argument := &gtype.AccountPasswordReset{}
	err := ctx.GetJson(&argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	account := strings.TrimSpace(argument.Account)
	if len(account) < 1 {
		ctx.Error(gtype.ErrInput, "帐号为空")
		return
	}
	site := &s.cfg.Site.Opt
	user := site.GetUser(account)
	if user == nil {
		ctx.Error(gtype.ErrNotExist, fmt.Sprintf("帐号(%s)不存在", account))
		return
	}
	password := strings.TrimSpace(argument.Password)

	cfg, err := s.cfg.Load()
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Errorf("load config fail: %s", err.Error()))
		return
	}
	cfgUser := cfg.Site.Opt.GetUser(account)
	if cfgUser == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("帐号(%s)不存在", account))
		return
	}
	cfgUser.Password = password
	err = s.cfg.Save(cfg)
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Errorf("save config fail: %s", err.Error()))
		return
	}

	user.Password = password
	ctx.Success(nil)
}

func (s *User) ResetPasswordDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "用户管理")
	function := catalog.AddFunction(method, uri, "重置本地用户密码")
	function.SetNote("重置用户的登录密码,内置管理员才能操作")
	function.SetInputJsonExample(&gtype.AccountPasswordReset{
		Account: "zs",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrNoPermission)
	function.AddOutputError(gtype.ErrExist)
}

func (s *User) ChangePassword(ctx gtype.Context, ps gtype.Params) {
	if s.cfg == nil {
		ctx.Error(gtype.ErrInternal, "cfg is nil")
		return
	}
	if s.cfg.Load == nil {
		ctx.Error(gtype.ErrInternal, "load not config")
		return
	}
	if s.cfg.Save == nil {
		ctx.Error(gtype.ErrInternal, "save not config")
		return
	}

	token := s.getToken(ctx.Token())
	if token == nil {
		ctx.Error(gtype.ErrInternal, "凭证无效")
		return
	}
	argument := &gtype.AccountPasswordChange{}
	err := ctx.GetJson(&argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Account) < 1 {
		argument.Account = token.UserAccount
	}
	account := strings.TrimSpace(argument.Account)
	if len(account) < 1 {
		ctx.Error(gtype.ErrInput, "帐号为空")
		return
	}
	site := &s.cfg.Site.Opt
	user := site.GetUser(account)
	if user == nil {
		ctx.Error(gtype.ErrNotExist, fmt.Sprintf("帐号(%s)不存在", account))
		return
	}
	oldPassword := strings.TrimSpace(argument.OldPassword)

	cfg, err := s.cfg.Load()
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Errorf("load config fail: %s", err.Error()))
		return
	}
	cfgUser := cfg.Site.Opt.GetUser(account)
	if cfgUser == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("帐号(%s)不存在", account))
		return
	}
	if cfgUser.Password != oldPassword {
		ctx.Error(gtype.ErrInput, "原密码错误")
		return
	}
	cfgUser.Password = strings.TrimSpace(argument.NewPassword)
	err = s.cfg.Save(cfg)
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Errorf("save config fail: %s", err.Error()))
		return
	}

	user.Password = cfgUser.Password
	ctx.Success(nil)
}

func (s *User) ChangePasswordDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "用户管理")
	function := catalog.AddFunction(method, uri, "修改本地用户密码")
	function.SetNote("修改用户的登录密码")
	function.SetInputJsonExample(&gtype.AccountPasswordChange{
		Account: "zs",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrNoPermission)
	function.AddOutputError(gtype.ErrExist)
}
