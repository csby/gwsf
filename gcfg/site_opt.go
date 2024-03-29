package gcfg

import (
	"github.com/csby/gwsf/gtype"
	"strings"
)

type SiteOpt struct {
	Path  string         `json:"path" note:"网站物理根路径, 如: /home/opt"`
	Api   SiteOptApi     `json:"api" note:"接口"`
	Users []*SiteOptUser `json:"users" note:"用户"`
	Ldap  SiteOptLdap    `json:"ldap" note:"LDAP验证"`

	DownloadTitle string `json:"downloadTitle" note:"下载连接标题"`
	DownloadUrl   string `json:"downloadUrl" note:"下载连接地址"`

	AccountVerification func(account, password string) gtype.Error `json:"-" note:"帐号密码验证"`
}

func (s *SiteOpt) GetUser(account string) *SiteOptUser {
	act := strings.ToLower(account)

	c := len(s.Users)
	for i := 0; i < c; i++ {
		u := s.Users[i]
		if u == nil {
			continue
		}
		if act == strings.ToLower(u.Account) {
			return u
		}
	}

	return nil
}

func (s *SiteOpt) RemoveUser(account string) int {
	act := strings.ToLower(account)

	users := make([]*SiteOptUser, 0)
	c := len(s.Users)
	for i := 0; i < c; i++ {
		u := s.Users[i]
		if u == nil {
			continue
		}
		if act == strings.ToLower(u.Account) {
			continue
		}

		users = append(users, u)
	}

	mc := c - len(users)
	if mc > 0 {
		s.Users = users
	}

	return mc
}
