package gcfg

type ServiceOther struct {
	Name        string `json:"name" note:"项目名称"`
	ServiceName string `json:"serviceName" note:"服务名称"`
	Remark      string `json:"remark" note:"备注说明"`
}
