package gcfg

type Service struct {
	Enabled bool             `json:"enabled" note:"是否启用(是否显示页面)"`
	Custom  ServiceCustom    `json:"custom" note:"自定义"`
	Tomcats []*ServiceTomcat `json:"tomcats" note:"tomcat"`
	Others  []*ServiceOther  `json:"others" note:"其他""`
	Nginxes []*ServiceNginx  `json:"nginxes" note:"nginx"`
	Files   []*ServiceFile   `json:"files" note:"文件"`
}

func (s *Service) GetTomcatByServiceName(name string) *ServiceTomcat {
	c := len(s.Tomcats)
	for i := 0; i < c; i++ {
		item := s.Tomcats[i]
		if item == nil {
			continue
		}

		if name == item.ServiceName {
			return item
		}
	}

	return nil
}

func (s *Service) GetOtherByServiceName(name string) *ServiceOther {
	c := len(s.Others)
	for i := 0; i < c; i++ {
		item := s.Others[i]
		if item == nil {
			continue
		}

		if name == item.ServiceName {
			return item
		}
	}

	return nil
}

func (s *Service) GetNginxByServiceName(name string) *ServiceNginx {
	c := len(s.Nginxes)
	for i := 0; i < c; i++ {
		item := s.Nginxes[i]
		if item == nil {
			continue
		}

		if name == item.ServiceName {
			return item
		}
	}

	return nil
}
