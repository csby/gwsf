package gmodel

type ServiceNginxLocation struct {
	Name       string   `json:"name" note:"名称"`
	Root       string   `json:"root" note:"根目录"`
	Urls       []string `json:"urls" note:"访问地址"`
	Version    string   `json:"version" note:"版本号"`
	DeployTime string   `json:"deployTime" note:"发布时间"`
}
