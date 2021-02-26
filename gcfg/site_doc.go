package gcfg

type SiteDoc struct {
	Path    string `json:"path" note:"网站物理根路径, 如: /home/doc"`
	Enabled bool   `json:"enabled" note:"是否启用"`

	DownloadTitle string `json:"downloadTitle" note:"下载连接标题"`
	DownloadUrl   string `json:"downloadUrl" note:"下载连接地址"`
}
