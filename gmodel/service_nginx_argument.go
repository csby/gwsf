package gmodel

type ServiceNginxArgument struct {
	SvcName    string `json:"svcName" note:"服务名称"`
	AppName    string `json:"appName" note:"站点名称"`
	Version    string `json:"version" note:"版本号"`
	DeployTime string `json:"deployTime" note:"发布时间"`
}

type ServiceNginxDetailArgument struct {
	SvcName string `json:"svcName" required:"true" note:"服务名称"`
	AppName string `json:"appName" required:"true" note:"站点名称"`
}
