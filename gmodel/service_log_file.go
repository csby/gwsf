package gmodel

import (
	"github.com/csby/gwsf/gtype"
	"sort"
	"strings"
)

type ServiceLogFile struct {
	Name     string                   `json:"name" note:"文件名称"`
	Size     int64                    `json:"size" note:"大小, 单位字节"`
	SizeText string                   `json:"sizeText" note:"大小文本信息"`
	ModTime  gtype.DateTime           `json:"modTime" note:"修改时间"`
	Path     string                   `json:"path" note:"路径, base64"`
	Folder   bool                     `json:"folder" note:"是否是文件夹"`
	Children ServiceLogFileCollection `json:"children" note:"子项"`
}

func (s *ServiceLogFile) Sort() {
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

type ServiceLogFileCollection []*ServiceLogFile

func (s ServiceLogFileCollection) Len() int {
	return len(s)
}

func (s ServiceLogFileCollection) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ServiceLogFileCollection) Less(i, j int) bool {
	if s[i].Folder {
		if s[j].Folder {
			return strings.Compare(s[i].Name, s[j].Name) < 0
		} else {
			return true
		}
	} else {
		return s[i].ModTime.After(s[j].ModTime)
	}
}
