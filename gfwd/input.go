package gfwd

import "github.com/csby/gwsf/gcfg"

type Input struct {
	forward

	InstanceID string
	Local      gcfg.Fwd
}
