package gcluster

import (
	"crypto/tls"
	"github.com/csby/gsecurity/gcrt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

type Certificate struct {
	gtype.Base

	ca  *gcrt.Crt
	crt *gcrt.Pfx
}

func (s *Certificate) Ca() *gcrt.Crt {
	return s.ca
}

func (s *Certificate) Crt() *gcrt.Pfx {
	return s.crt
}

func (s *Certificate) Load(instance *gcfg.ClusterInstance) {
	if instance == nil {
		return
	}

	caFilePath := instance.Ca.File
	s.LogInfo("cluster ca file: ", caFilePath)
	pfxFilePath := instance.Crt.File
	s.LogInfo("cluster crt file: ", pfxFilePath)
	pfx := &gcrt.Pfx{}
	err := pfx.FromFile(pfxFilePath, instance.Crt.Password)
	if err != nil {
		s.crt = nil
		s.LogError("load cluster crt file fail: ", err)
	} else {
		s.crt = pfx
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
	} else {
		cfg.InsecureSkipVerify = true
	}
	if s.crt != nil {
		cfg.Certificates = s.crt.TlsCertificates()
	}

	return cfg
}
