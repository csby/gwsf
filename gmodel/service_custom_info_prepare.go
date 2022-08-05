package gmodel

type ServiceCustomInfoPrepare struct {
	Exec string `json:"exec" note:"可执行程序"`
	Args string `json:"args" note:"程序启动参数"`
}
