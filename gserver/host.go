package gserver

import (
	"crypto/tls"
	"fmt"
	"github.com/csby/gsecurity/gcrt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"net"
	"net/http"
	"strings"
	"sync"
)

type host struct {
	gtype.Base

	cfg         *gcfg.Config
	httpHandler gtype.Handler
	httpServer  *http.Server
	httpsServer *http.Server
}

func (s *host) Run() error {
	if s.cfg == nil {
		return fmt.Errorf(s.LogError("invalid configure: nil"))
	}

	wg := &sync.WaitGroup{}

	if s.httpHandler != nil {
		router, err := newHandler(s.GetLog(), s.cfg, s.httpHandler)
		if err != nil {
			s.LogError("newHandler error: ", err)
			return err
		}

		// http
		if s.cfg.Http.Enabled {
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer s.LogInfo("http server stopped")

				err := s.runHttp(router)
				if err != nil {
					s.LogError("http server error: ", err)
				}

			}()
		}

		// https
		if s.cfg.Https.Enabled {
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer s.LogInfo("https server stopped")

				err := s.runHttps(router)
				if err != nil {
					s.LogError("https server error: ", err)
				}
			}()
		}
	}

	wg.Wait()
	return nil
}

func (s *host) Close() (err error) {
	if s.httpServer != nil {
		e := s.httpServer.Close()
		if e != nil {
			err = e
		}
	}

	if s.httpsServer != nil {
		e := s.httpsServer.Close()
		if e != nil {
			err = e
		}
	}

	return
}

func (s *host) runHttp(handler *handler) error {
	defer func() {
		if err := recover(); err != nil {
			s.LogError("http server exception: ", err)
		}
	}()

	addr := fmt.Sprintf("%s:%d", s.cfg.Http.Address, s.cfg.Http.Port)
	s.LogInfo("http server running on \"", addr, "\"")

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	if s.cfg.Http.BehindProxy {
		s.httpServer.ProxyRemoteAddr = s.getRemoteAddr
	}
	err := s.httpServer.ListenAndServe()
	s.httpServer = nil

	return err
}

func (s *host) runHttps(handler *handler) error {
	defer func() {
		if err := recover(); err != nil {
			s.LogError("https server exception: ", err)
		}
	}()

	caFilePath := s.cfg.Https.Cert.Ca.File
	s.LogInfo("https server ca file: ", caFilePath)
	pfxFilePath := s.cfg.Https.Cert.Server.File
	s.LogInfo("https server pfx file: ", pfxFilePath)
	pfx := &gcrt.Pfx{}
	err := pfx.FromFile(pfxFilePath, s.cfg.Https.Cert.Server.Password)
	if err != nil {
		return fmt.Errorf("load pfx file fail: %v", err)
	}
	handler.serverCrt = pfx

	addr := fmt.Sprintf("%s:%d", s.cfg.Https.Address, s.cfg.Https.Port)
	s.httpsServer = &http.Server{
		Addr:    addr,
		Handler: handler,
		TLSConfig: &tls.Config{
			Certificates: pfx.TlsCertificates(),
			ClientAuth:   tls.NoClientCert,
		},
	}
	if s.cfg.Https.BehindProxy {
		s.httpsServer.ProxyRemoteAddr = s.getRemoteAddr
	}
	if s.cfg.Https.RequestClientCert {
		s.httpsServer.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
	if len(caFilePath) > 0 {
		crt := &gcrt.Crt{}
		err = crt.FromFile(caFilePath)
		if err != nil {
			return fmt.Errorf("load ca file fail: %v", err)
		}
		handler.caCrt = crt
		s.httpsServer.TLSConfig.ClientCAs = crt.Pool()
	}

	s.LogInfo("https server running on \"", addr, "\"")
	err = s.httpsServer.ListenAndServeTLS("", "")
	s.httpsServer = nil

	return err
}

func (s *host) getRemoteAddr(conn net.Conn) string {
	if len(s.cfg.Proxy) > 0 {
		addr := fmt.Sprint(conn.RemoteAddr())
		ip, _, _ := net.SplitHostPort(addr)
		if len(ip) > 0 {
			if ip != s.cfg.Proxy {
				return addr
			}
		}
	}

	rawConn := conn
	if tlsConn, ok := conn.(*tls.Conn); ok {
		rawConn = tlsConn.RawConn()
	}

	buf := make([]byte, 1)
	sb := &strings.Builder{}
	for {
		_, e := rawConn.Read(buf)
		if e != nil {
			return ""
		}

		if buf[0] == '\n' {
			break
		}
		if buf[0] == '\r' {
			continue
		}

		sb.Write(buf)
	}

	// PROXY family srcIP srcPort targetIP targetPort
	// PROXY TCP4 192.168.123.254 12955 192.168.123.81 8088
	proxy := sb.String()
	vs := strings.Split(proxy, " ")
	if len(vs) > 3 {
		return fmt.Sprintf("%s:%s", vs[2], vs[3])
	} else {
		return ""
	}
}
