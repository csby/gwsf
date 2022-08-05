package gmodel

type ServerArgument struct {
	Name string `json:"name" required:"true" note:"名称"`
}
