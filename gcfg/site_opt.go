package gcfg

type SiteOpt struct {
	Path  string        `json:"path" note:"网站物理根路径, 如: /home/opt"`
	Api   SiteOptApi    `json:"api" note:"接口"`
	Users []SiteOptUser `json:"users" note:"用户"`
	Ldap  SiteOptLdap   `json:"ldap" note:"LDAP验证"`
}
