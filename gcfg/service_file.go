package gcfg

type ServiceFile struct {
	Path    string `json:"path" note:"路径, 如： /download"`
	Root    string `json:"root" note:"根目录"`
	Enabled bool   `json:"enabled" note:"是否启用"`
}
