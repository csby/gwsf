package gmodel

import (
	"encoding/json"
	"fmt"
	"github.com/csby/gwsf/gtype"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ServiceCustomShellInfo struct {
	Name       string         `json:"name" note:"名称"`
	Version    string         `json:"version" note:"版本号"`
	DeployTime gtype.DateTime `json:"deployTime" note:"发布时间"`
}

type ServiceCustomInfo struct {
	sync.RWMutex

	Name string `json:"name" note:"项目名称"`
	Exec string `json:"exec" note:"可执行程序"`
	Args string `json:"args" note:"程序启动参数"`

	ServiceName string                      `json:"serviceName" note:"服务名称"`
	DisplayName string                      `json:"displayName" note:"显示名称"`
	Description string                      `json:"description" note:"描述信息"`
	Remark      string                      `json:"remark" note:"备注信息"`
	Version     string                      `json:"version" note:"版本号"`
	Author      string                      `json:"author" note:"作者"`
	Folder      string                      `json:"folder" note:"物理目录"`
	DeployTime  gtype.DateTime              `json:"deployTime" note:"发布时间"`
	Status      gtype.ServerStatus          `json:"status" note:"状态: 0-未安装; 1-运行中; 2-已停止"`
	Prepares    []*ServiceCustomInfoPrepare `json:"prepares" note:"预执行程序(主程序运行前执行)"`
}

func (s *ServiceCustomInfo) LoadFromFile(filePath string) error {
	s.Lock()
	defer s.Unlock()

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, s)
}

func (s *ServiceCustomInfo) SaveToFile(filePath string) error {
	s.Lock()
	defer s.Unlock()

	bytes, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return err
	}

	fileFolder := filepath.Dir(filePath)
	_, err = os.Stat(fileFolder)
	if os.IsNotExist(err) {
		os.MkdirAll(fileFolder, 0777)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprint(file, string(bytes[:]))

	return err
}

func (s *ServiceCustomInfo) GetServiceName() string {
	return fmt.Sprintf("svc-cst-%s", s.Name)
}

type ServiceCustomInfoCollection []*ServiceCustomInfo

func (s ServiceCustomInfoCollection) Len() int {
	return len(s)
}

func (s ServiceCustomInfoCollection) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ServiceCustomInfoCollection) Less(i, j int) bool {
	return strings.Compare(s[i].Name, s[j].Name) < 0
}
