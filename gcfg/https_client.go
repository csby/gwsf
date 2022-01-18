package gcfg

import "github.com/csby/gwsf/gtype"

type HttpsClient struct {
	Id          string             `json:"id" note:"证书标识ID, 对于证书OU"`
	Name        string             `json:"name" note:"证书名称, 对于证书CN"`
	Kind        string             `json:"kind" note:"证书类型, 对于证书O"`
	Addr        HttpsClientAddress `json:"addr" note:"地址"`
	RegTime     gtype.DateTime     `json:"regTime" note:"注册时间"`
	RegIp       string             `json:"regIp" note:"注册IP地址"`
	OnlineTime  gtype.DateTime     `json:"onlineTime" note:"最近上线时间"`
	OfflineTime *gtype.DateTime    `json:"offlineTime" note:"最近离线时间"`

	DisplayName string `json:"displayName" note:"显示名称"`
	Remark      string `json:"remark" note:"备注信息"`
}

type HttpsClientAddress struct {
	Province string `json:"province" note:"省份, 对于证书S"`
	Locality string `json:"locality" note:"地区, 对于证书L"`
	Address  string `json:"address" note:"地址, 对于证书STREET"`
}
