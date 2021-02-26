package gproxy

type Status int

const (
	StatusStopped  Status = 0
	StatusStarting Status = 1
	StatusRunning  Status = 2
	StatusStopping Status = 3
)

var statuses = [...]string{
	"stopped",
	"starting",
	"running",
	"stopping",
}

func (s Status) String() string {
	if s >= StatusStopped && s <= StatusStopping {
		return statuses[s]
	}

	return ""
}
