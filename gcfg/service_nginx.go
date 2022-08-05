package gcfg

type ServiceNginx struct {
	Name        string `json:"name" note:"项目名称"`
	ServiceName string `json:"serviceName" note:"服务名称"`
	Remark      string `json:"remark" note:"备注说明"`
	Log         string `json:"log" note:"日志根目录"`

	Locations []*ServiceNginxLocation `json:"locations" note:"位置"`
}

func (s *ServiceNginx) GetLocationByName(name string) *ServiceNginxLocation {
	c := len(s.Locations)
	for i := 0; i < c; i++ {
		item := s.Locations[i]
		if item == nil {
			continue
		}

		if name == item.Name {
			return item
		}
	}

	return nil
}
