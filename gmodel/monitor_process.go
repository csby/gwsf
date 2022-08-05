package gmodel

type ProcessID struct {
	Pid int `json:"pid" required:"true" note:"进程ID"`
}
