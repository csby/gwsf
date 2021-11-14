package gcfg

import "github.com/csby/gwsf/gtype"

type SiteOpt struct {
	Path  string        `json:"path" note:"网站物理根路径, 如: /home/opt"`
	Api   SiteOptApi    `json:"api" note:"接口"`
	Users []SiteOptUser `json:"users" note:"用户"`
	Ldap  SiteOptLdap   `json:"ldap" note:"LDAP验证"`

	DownloadTitle string `json:"downloadTitle" note:"下载连接标题"`
	DownloadUrl   string `json:"downloadUrl" note:"下载连接地址"`

	AccountVerification func(account, password string) gtype.Error `json:"-" note:"帐号密码验证"`
}
