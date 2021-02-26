package gtype

type WebAppId struct {
	Id string `json:"id" note:"标识ID" required:"true"`
}

type WebApp struct {
	WebAppId

	Name       string    `json:"name" note:"网站名称，如： 文档中心"`
	Url        string    `json:"url" note:"访问地址"`
	Root       string    `json:"root" note:"物理路径"`
	Guid       string    `json:"guid" note:"网站ID,更新网站时将对该值进行校验,在文件site.json文件中"`
	Version    string    `json:"version" note:"版本号,在文件site.json文件中"`
	DeployTime *DateTime `json:"deployTime" note:"发布时间"`

	DownloadTitle string `json:"downloadTitle" note:"下载连接标题"`
	DownloadUrl   string `json:"downloadUrl" note:"下载连接地址"`
}
