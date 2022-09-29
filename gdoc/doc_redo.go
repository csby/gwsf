package gdoc

import "github.com/csby/gwsf/gtype"

type redo struct {
	method string
	uri    gtype.Uri
	handle gtype.DocHandle
}
