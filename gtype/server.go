package gtype

type Server interface {
	ServiceName() string
	Interactive() bool
	Run() error
	Shutdown() error
	Restart() error
	Start() error
	Stop() error
	Install() error
	Uninstall() error
	Status() (ServerStatus, error)
}

type ServerStatus byte

const (
	ServerStatusUnknown ServerStatus = iota
	ServerStatusRunning
	ServerStatusStopped
)

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerRole struct {
	Cluster bool `json:"cluster" note:"是否集群"`
	Cloud   bool `json:"cloud" note:"是否云端"`
	Node    bool `json:"node" note:"是否节点"`
	Service bool `json:"service" note:"是否服务管理"`
	Proxy   bool `json:"proxy" note:"是否反向代理"`
}
