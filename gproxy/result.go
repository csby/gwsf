package gproxy

import "github.com/csby/gwsf/gtype"

type Result struct {
	Status    Status          `json:"status" note:"服务状态：0-已停止; 1-启动中; 2-运行中; 3-停止中"`
	StartTime *gtype.DateTime `json:"startTime" note:"启动时间"`
	Error     string          `json:"error" note:"错误信息"`
}
