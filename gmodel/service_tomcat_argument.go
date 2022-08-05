package gmodel

type ServiceTomcatArgument struct {
	Name string `json:"name" note:"服务名称"`
	App  string `json:"app" note:"应用名称"`
}
