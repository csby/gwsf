package gcfg

type SiteOptUser struct {
	Account  string `json:"account" note:"账号"`
	Password string `json:"password" note:"密码"`
	Name     string `json:"name" note:"姓名"`
}
