package gcfg

import "time"

type Svc struct {
	Name     string    `json:"name" note:"服务名称，系统内唯一"`
	Args     string    `json:"-" note:"启动参数"`
	BootTime time.Time `json:"-" note:"启动时间"`

	DownloadTitle string `json:"downloadTitle" note:"下载连接标题"`
	DownloadUrl   string `json:"downloadUrl" note:"下载连接地址"`

	Restart func() error `json:"-" note:"重启服务"`
}
