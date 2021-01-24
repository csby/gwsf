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
