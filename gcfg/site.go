package gcfg

type Site struct {
	Root SiteRoot  `json:"root" note:"根站点"`
	Doc  SiteDoc   `json:"doc" note:"文档网站"`
	Opt  SiteOpt   `json:"opt" note:"管理网站"`
	Apps []SiteApp `json:"apps" note:"应用网站"`
}
