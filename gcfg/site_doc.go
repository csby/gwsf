package gcfg

type SiteDoc struct {
	Path    string `json:"path" note:"网站物理根路径, 如: /home/doc"`
	Enabled bool   `json:"enabled" note:"是否启用"`
}
