package gmodel

type ServiceLogArgument struct {
	SvcName  string `json:"svcName" required:"true" not:"服务名称"`
	FileName string `json:"fileName" required:"true" not:"文件名称"`
}
