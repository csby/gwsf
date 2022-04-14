package gtype

import "strings"

type DatabaseServer struct {
	Host string `json:"host" required:"true" note:"主机"`
	Port string `json:"port" note:"端口, 默认1434"`
}

type SqlServerInstance struct {
	Name    string `json:"name" note:"实例名称"`
	Port    string `json:"port" note:"实例监听端口"`
	Version string `json:"version" note:"实例版本号"`
}

type SqlServerInstanceCollection []*SqlServerInstance

func (s SqlServerInstanceCollection) Len() int      { return len(s) }
func (s SqlServerInstanceCollection) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SqlServerInstanceCollection) Less(i, j int) bool {
	return strings.Compare(s[i].Name, s[j].Name) < 0
}
