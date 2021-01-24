package gcfg

type Site struct {
	Root SiteRoot `json:"root" note:"根站点"`
	Doc  SiteDoc  `json:"doc" note:"文档网站"`
}
