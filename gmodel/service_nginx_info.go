package gmodel

import "github.com/csby/gwsf/gtype"

type ServiceNginxInfo struct {
	Name        string             `json:"name" note:"项目名称"`
	ServiceName string             `json:"serviceName" note:"服务名称"`
	Remark      string             `json:"remark" note:"备注说明"`
	Status      gtype.ServerStatus `json:"status" note:"状态: 0-未安装; 1-运行中; 2-已停止"`

	Locations []*ServiceNginxLocation `json:"locations" note:"位置"`
}
