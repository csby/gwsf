package gcfg

type ServiceNginxLocation struct {
	Name string   `json:"name" note:"名称"`
	Root string   `json:"root" note:"根目录"`
	Urls []string `json:"urls" note:"访问地址"`
}
