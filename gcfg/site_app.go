package gcfg

type SiteApp struct {
	Name string `json:"name" note:"网站名称，如： 文档中心"`
	Path string `json:"path" note:"网站物理根路径, 如: /home/apps/test"`
	Uri  string `json:"uri" note:"基本URL，如 /doc-center"`
}
