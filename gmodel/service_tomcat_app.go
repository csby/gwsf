package gmodel

import (
	"github.com/csby/gwsf/gtype"
	"strings"
)

type ServiceTomcatApp struct {
	Name       string         `json:"name" note:"名称"`
	Version    string         `json:"version" note:"版本号"`
	DeployTime gtype.DateTime `json:"deployTime" note:"发布时间"`
}

type ServiceTomcatAppCollection []*ServiceTomcatApp

func (s ServiceTomcatAppCollection) Len() int {
	return len(s)
}

func (s ServiceTomcatAppCollection) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ServiceTomcatAppCollection) Less(i, j int) bool {
	return strings.Compare(s[i].Name, s[j].Name) < 0
}
