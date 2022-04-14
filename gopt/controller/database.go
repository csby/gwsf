package controller

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

func NewDatabase(log gtype.Log, cfg *gcfg.Config) *Database {
	instance := &Database{}
	instance.SetLog(log)
	instance.cfg = cfg

	return instance
}

type Database struct {
	controller
}
