package gnode

import (
	"crypto/tls"
	"github.com/csby/gsecurity/gcrt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

type Certificate struct {
	gtype.Base

	ca   *gcrt.Crt
	node *gcrt.Pfx
}

func (s *Certificate) Ca() *gcrt.Crt {
	return s.ca
}

func (s *Certificate) Node() *gcrt.Pfx {
	return s.node
}

func (s *Certificate) Load(cfg *gcfg.Crt) {
	if cfg == nil {
		return
	}

	caFilePath := cfg.Ca.File
	s.LogInfo("node ca file: ", caFilePath)
	pfxFilePath := cfg.Server.File
	s.LogInfo("node pfx file: ", pfxFilePath)
	pfx := &gcrt.Pfx{}
	err := pfx.FromFile(pfxFilePath, cfg.Server.Password)
	if err != nil {
		s.node = nil
		s.LogError("load node pfx file fail: ", err)
	} else {
		s.node = pfx
	}

	if len(caFilePath) > 0 {
		crt := &gcrt.Crt{}
		err = crt.FromFile(caFilePath)
		if err != nil {
			s.ca = nil
			s.LogError("load node ca file fail: ", err)
		} else {
			s.ca = crt
		}
	} else {
		s.ca = nil
	}
}

func (s *Certificate) ClientTlsConfig() *tls.Config {
	cfg := &tls.Config{}

	if s.ca != nil {
		cfg.RootCAs = s.ca.Pool()
	}
	if s.node != nil {
		cfg.Certificates = s.node.TlsCertificates()
	}

	return cfg
}
