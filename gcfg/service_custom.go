package gcfg

type ServiceCustom struct {
	App           string `json:"app" note:"程序根目录"`
	Log           string `json:"log" note:"日志根目录"`
	LogRetainDays int64  `json:"logRetainDays" note:"日志保留天数, 0表示永久保留"`
}
