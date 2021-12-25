package gproxy

import (
	"testing"
)

func TestServer_IsAlive(t *testing.T) {
	s := &Server{}
	err := s.isAlive("192.168.1.1:443")
	t.Log(err)
}
