package gmodel

import "github.com/csby/gwsf/gtype"

type ServiceStatus struct {
	Name   string             `json:"name" not:"名称"`
	Status gtype.ServerStatus `json:"status" note:"状态: 0-未安装; 1-运行中; 2-已停止"`
}
