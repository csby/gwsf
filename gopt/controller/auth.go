package controller

import (
	"encoding/base64"
	"fmt"
	"github.com/csby/gsecurity/grsa"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"github.com/mojocn/base64Captcha"
	"strings"
	"time"
)

const captchaNumberSource = "1234567890"
const captchaLetterSource = "ABCDEFGHIJKLMNPQRSTUVWXYZabcedefghijklmnpqrstuvwxyz"
const captchaLetterNumberSource = "ABCDEFGHIJKLMNPQRSTUVWXYZ123456789abcedefghijklmnpqrstuvwxyz"

type Auth struct {
	controller

	errorCount   map[string]int
	ldap         *Ldap
	captchaStore base64Captcha.Store
	rsaPrivate   grsa.Private

	AccountVerification func(account, password string) gtype.Error
}

func NewAuth(log gtype.Log, cfg *gcfg.Config, db gtype.TokenDatabase, chs gtype.SocketChannelCollection) *Auth {
	instance := &Auth{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.dbToken = db
	instance.wsChannels = chs
	instance.errorCount = make(map[string]int)
	instance.captchaStore = base64Captcha.DefaultMemStore
	instance.rsaPrivate.Create(1024)

	instance.ldap = &Ldap{}
	if cfg != nil {
		instance.ldap.init(&cfg.Site.Opt.Ldap)
	}

	if chs != nil {
		chs.AddFilter(instance.onWebsocketWriteFilter)
	}

	return instance
}

func (s *Auth) GetCaptcha(ctx gtype.Context, ps gtype.Params) {
	filter := &gtype.CaptchaFilter{
		Mode:   3,
		Length: 4,
		Width:  100,
		Height: 30,
	}
	err := ctx.GetJson(filter)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	noiseCount := 0
	showLineOptions := 0
	var driver base64Captcha.Driver
	switch filter.Mode {
	case 0:
		driver = base64Captcha.NewDriverString(filter.Height, filter.Width, noiseCount, showLineOptions, filter.Length, captchaNumberSource, &filter.BackColor, nil, []string{})
	case 1:
		driver = base64Captcha.NewDriverString(filter.Height, filter.Width, noiseCount, showLineOptions, filter.Length, captchaLetterSource, &filter.BackColor, nil, []string{})
	case 2:
		driver = base64Captcha.NewDriverMath(filter.Height, filter.Width, noiseCount, showLineOptions, &filter.BackColor, nil, []string{})
	default:
		driver = base64Captcha.NewDriverString(filter.Height, filter.Width, noiseCount, showLineOptions, filter.Length, captchaLetterNumberSource, &filter.BackColor, nil, []string{})
	}

	captcha := base64Captcha.NewCaptcha(driver, s.captchaStore)
	captchaId, captchaValue, err := captcha.Generate()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	data := &gtype.Captcha{
		ID:       captchaId,
		Value:    captchaValue,
		Required: s.captchaRequired(ctx.RIP()),
	}
	rsaPublic, err := s.rsaPrivate.Public()
	if err == nil {
		keyVal, err := rsaPublic.ToMemory()
		if err == nil {
			data.RsaPublicKey = string(keyVal)
		}
	}

	ctx.Success(data)
}

func (s *Auth) GetCaptchaDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "权限管理")
	function := catalog.AddFunction(method, uri, "获取验证码")
	function.SetNote("获取用户登陆需要的验证码信息")
	function.SetRemark("该接口不需要凭证")
	function.SetInputJsonExample(&gtype.CaptchaFilter{
		Mode:   3,
		Length: 4,
		Width:  100,
		Height: 30,
	})

	function.SetOutputDataExample(&gtype.Captcha{
		ID:           "GKSVhVMRAHsyVuXSrMYs",
		Value:        "data:image/png;base64,iVBOR...",
		RsaPublicKey: "-----BEGIN PUBLIC KEY-----...-----END PUBLIC KEY-----",
		Required:     false,
	})
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
}

func (s *Auth) Login(ctx gtype.Context, ps gtype.Params) {
	filter := &gtype.LoginFilter{}
	err := ctx.GetJson(filter)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	requireCaptcha := s.captchaRequired(ctx.RIP())
	err = filter.Check(requireCaptcha)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	if requireCaptcha {
		captchaValue := s.captchaStore.Get(filter.CaptchaId, true)
		if strings.ToLower(captchaValue) != strings.ToLower(filter.CaptchaValue) {
			ctx.Error(gtype.ErrLoginCaptchaInvalid)
			return
		}
	}

	pwd := filter.Password
	if strings.ToLower(filter.Encryption) == "rsa" {
		buf, err := base64.StdEncoding.DecodeString(filter.Password)
		if err != nil {
			ctx.Error(gtype.ErrLoginPasswordInvalid, err)
			s.increaseErrorCount(ctx.RIP())
			return
		}

		decryptedPwd, err := s.rsaPrivate.Decrypt(buf)
		if err != nil {
			ctx.Error(gtype.ErrLoginPasswordInvalid, err)
			s.increaseErrorCount(ctx.RIP())
			return
		}
		pwd = string(decryptedPwd)
	}

	login, be, err := s.Authenticate(ctx, filter.Account, pwd)
	if be != nil {
		ctx.Error(be, err)
		s.increaseErrorCount(ctx.RIP())
		return
	}

	ctx.Success(login)
	s.clearErrorCount(ctx.RIP())
}

func (s *Auth) LoginDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "权限管理")
	function := catalog.AddFunction(method, uri, "用户登录")
	function.SetNote("通过用户账号及密码进行登录获取凭证")
	function.SetRemark("连续3次错误将要求输入验证码")
	function.SetInputJsonExample(&gtype.LoginFilter{
		Account:      "admin",
		Password:     "1",
		CaptchaId:    "r4kcmz2E12e0qJQOvqRB",
		CaptchaValue: "1e35",
		Encryption:   "",
	})

	function.SetOutputDataExample(&gtype.Login{
		Token: "71b9b7e2ac6d4166b18f414942ff3481",
	})
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrLoginCaptchaInvalid)
	function.AddOutputError(gtype.ErrLoginAccountNotExit)
	function.AddOutputError(gtype.ErrLoginPasswordInvalid)
	function.AddOutputError(gtype.ErrLoginAccountOrPasswordInvalid)
}

func (s *Auth) Logout(ctx gtype.Context, ps gtype.Params) {
	tv := ctx.Token()
	if len(tv) < 1 {
		ctx.Error(gtype.ErrTokenEmpty)
		return
	}
	_, ok := s.dbToken.Get(tv, false)
	if !ok {
		ctx.Error(gtype.ErrTokenInvalid)
		return
	}

	s.writeWebSocketMessage(ctx.Token(), gtype.WSOptUserLogout, tv)
	ok = s.dbToken.Del(tv)
	if ok {
	}

	ctx.Success(nil)
}

func (s *Auth) LogoutDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "权限管理")
	function := catalog.AddFunction(method, uri, "退出登录")
	function.SetNote("退出登录, 使当前凭证失效")
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Auth) SetLdap(ctx gtype.Context, ps gtype.Params) {
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
		ctx.Error(gtype.ErrNoPermission, fmt.Sprintf("只有内置管理员帐号(%s)才能进行操作", adminAccount))
		return
	}

	argument := &gcfg.SiteOptLdap{}
	err := ctx.GetJson(&argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	cfg, err := s.cfg.Load()
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Errorf("load config fail: %s", err.Error()))
		return
	}
	count := argument.CopyTo(&cfg.Site.Opt.Ldap)
	if count < 1 {
		ctx.Success(count)
		return
	}
	err = s.cfg.Save(cfg)
	if err != nil {
		ctx.Error(gtype.ErrInternal, fmt.Errorf("save config fail: %s", err.Error()))
		return
	}

	argument.CopyTo(&s.cfg.Site.Opt.Ldap)
	s.ldap.init(argument)

	ctx.Success(count)
}

func (s *Auth) SetLdapDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "用户管理")
	function := catalog.AddFunction(method, uri, "设置LDAP")
	function.SetNote("修改LDAP配置,成功时返回被修改属性的数量")
	function.SetInputJsonExample(&gcfg.SiteOptLdap{
		Host: "192.168.1.1",
		Port: 389,
		Base: "dc=example,dc=com",
	})
	function.SetOutputDataExample(0)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrNoPermission)
	function.AddOutputError(gtype.ErrExist)
}

func (s *Auth) GetLdap(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(&s.cfg.Site.Opt.Ldap)
}

func (s *Auth) GetLdapDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "用户管理")
	function := catalog.AddFunction(method, uri, "获取LDAP")
	function.SetNote("获取LDAP配置")
	function.SetOutputDataExample(&gcfg.SiteOptLdap{
		Host: "192.168.1.1",
		Port: 389,
		Base: "dc=example,dc=com",
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Auth) CreateTokenForAccountPassword(items []gtype.TokenAuth, ctx gtype.Context) (string, gtype.Error) {
	account := ""
	password := ""
	count := len(items)
	for i := 0; i < count; i++ {
		item := items[i]
		if item.Name == "account" {
			account = item.Value
		} else if item.Name == "password" {
			password = item.Value
		}
	}

	model, code, err := s.Authenticate(ctx, account, password)
	if code != nil {
		return "", code.SetDetail(err)
	}

	if model != nil {
		return model.Token, nil
	}

	return "", nil
}

func (s *Auth) Authenticate(ctx gtype.Context, account, password string) (*gtype.Login, gtype.Error, error) {
	act := strings.ToLower(account)
	pwd := password
	userName := account

	var err error = nil
	if s.AccountVerification != nil {
		ge := s.AccountVerification(act, pwd)
		if ge != nil {
			return nil, ge, nil
		}
	} else {
		var user *gcfg.SiteOptUser = nil
		userCount := len(s.cfg.Site.Opt.Users)
		for index := 0; index < userCount; index++ {
			if act == strings.ToLower(s.cfg.Site.Opt.Users[index].Account) {
				user = s.cfg.Site.Opt.Users[index]
				break
			}
		}

		if user != nil {
			if pwd != user.Password {
				return nil, gtype.ErrLoginPasswordInvalid, nil
			}
			if len(user.Name) > 0 {
				userName = user.Name
			}
		} else {
			if s.ldap.Enable {
				err = s.ldap.Authenticate(account, password)
				if err != nil {
					return nil, gtype.ErrLoginAccountOrPasswordInvalid, err
				}
			} else {
				return nil, gtype.ErrLoginAccountNotExit, nil
			}
		}
	}

	now := time.Now()
	token := &gtype.Token{
		ID:          ctx.NewGuid(),
		UserAccount: account,
		UserName:    userName,
		LoginIP:     ctx.RIP(),
		LoginTime:   now,
		ActiveTime:  now,
		Usage:       0,
	}
	s.dbToken.Set(token.ID, token)

	login := &gtype.Login{
		Token:   token.ID,
		Account: token.UserAccount,
		Name:    token.UserName,
	}

	return login, nil, err
}

func (s *Auth) CheckToken(ctx gtype.Context, ps gtype.Params) {
	tokenValue := ctx.Token()
	if len(tokenValue) < 1 {
		ctx.Error(gtype.ErrTokenEmpty)
		ctx.SetHandled(true)
		return
	}

	token, ok := s.dbToken.Get(tokenValue, true)
	if !ok {
		ctx.Error(gtype.ErrTokenInvalid)
		ctx.SetHandled(true)
		return
	}

	tokenModel, ok := token.(*gtype.Token)
	if !ok {
		ctx.Error(gtype.ErrInternal, "类型转换错误(*gtype.Token)")
		ctx.SetHandled(true)
		return
	}

	if tokenModel.LoginIP != ctx.RIP() {
		ctx.Error(gtype.ErrTokenIllegal, fmt.Sprintf("IP不匹配: 当前IP%s, 登录IP%s", ctx.RIP(), tokenModel.LoginIP))
		ctx.SetHandled(true)
		return
	}

	ctx.Set(gtype.CtxUserAccount, tokenModel.UserAccount)
}

func (s *Auth) onWebsocketWriteFilter(message *gtype.SocketMessage, channel gtype.SocketChannel, token *gtype.Token) bool {
	if message == nil {
		return false
	}

	if channel == nil {
		return false
	}

	if token == nil {
		return false
	}
	channelToken := channel.Token()
	if channelToken == nil {
		return false
	}

	if message.ID == gtype.WSOptUserLogin {
		if channelToken.ID == token.ID {
			return true
		}
	} else if message.ID == gtype.WSOptUserLogout {
		if channelToken.ID != token.ID {
			return true
		}
	}

	return false
}

func (s *Auth) captchaRequired(ip string) bool {
	if s.errorCount == nil {
		return false
	}

	count, ok := s.errorCount[ip]
	if ok {
		if count < 3 {
			return false
		} else {
			return true
		}
	}

	return false
}

func (s *Auth) increaseErrorCount(ip string) {
	if s.errorCount == nil {
		return
	}

	count := 1
	v, ok := s.errorCount[ip]
	if ok {
		count += v
	}

	s.errorCount[ip] = count
}

func (s *Auth) clearErrorCount(ip string) {
	if s.errorCount == nil {
		return
	}

	_, ok := s.errorCount[ip]
	if ok {
		delete(s.errorCount, ip)
	}
}
