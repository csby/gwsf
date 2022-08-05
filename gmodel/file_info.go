package gmodel

import (
	"github.com/csby/gwsf/gtype"
	"sort"
	"strings"
)

type FileInfo struct {
	Name     string             `json:"name" note:"名称"`
	Path     string             `json:"path" note:"路径，base64"`
	Size     int64              `json:"size" note:"大小，字节数"`
	Time     gtype.DateTime     `json:"time" note:"修改时间"`
	Folder   bool               `json:"folder" note:"true-文件夹; false-文件"`
	Children FileInfoCollection `json:"children" note:"子项"`
}

func (s *FileInfo) Sort() {
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

type FileInfoCollection []*FileInfo

func (s FileInfoCollection) Len() int {
	return len(s)
}

func (s FileInfoCollection) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s FileInfoCollection) Less(i, j int) bool {
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

type FileInfoArgument struct {
	Path string `json:"path" required:"true" note:"路径，base64"`
}
