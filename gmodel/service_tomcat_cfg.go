package gmodel

import (
	"sort"
	"strings"
)

type ServiceTomcatCfg struct {
	Name     string                     `json:"name" note:"名称"`
	Path     string                     `json:"path" note:"路径，base64"`
	Folder   bool                       `json:"folder" note:"true-文件夹; false-文件"`
	Children ServiceTomcatCfgCollection `json:"children" note:"子项"`
}

func (s *ServiceTomcatCfg) Sort() {
	c := len(s.Children)
	if c < 2 {
		return
	}
	sort.Sort(s.Children)

	for i := 0; i < c; i++ {
		item := s.Children[i]
		if item == nil {
			continue
		}
		if !item.Folder {
			continue
		}

		item.Sort()
	}
}

type ServiceTomcatCfgCollection []*ServiceTomcatCfg

func (s ServiceTomcatCfgCollection) Len() int {
	return len(s)
}

func (s ServiceTomcatCfgCollection) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ServiceTomcatCfgCollection) Less(i, j int) bool {
	if s[i].Folder {
		if s[j].Folder {
			return strings.Compare(s[i].Name, s[j].Name) < 0
		} else {
			return true
		}
	} else {
		return strings.Compare(s[i].Name, s[j].Name) < 0
	}
}

type ServiceTomcatCfgFolder struct {
	Name   string `json:"name" note:"服务名称"`
	Path   string `json:"path" note:"路径，base64"`
	Folder string `json:"folder" note:"文件夹名称"`
}
