package gtype

import "github.com/csby/gsecurity/gcrt"

type Certificate struct {
	Ca     *gcrt.Crt
	Client *gcrt.Crt
	Server *gcrt.Pfx
}
