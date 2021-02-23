package test

import (
	"crypto/tls"
	"fmt"
	"github.com/csby/gsecurity/gcrt"
	"github.com/csby/gwsf/gclient"
	"github.com/csby/gwsf/gtype"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	wg        sync.WaitGroup
	transport *http.Transport
}

func (s *Client) Run() <-chan error {
	// load certificate for client
	clientCrt := &gcrt.Pfx{}
	err := clientCrt.FromFile(clientCrtFilePath, clientCrtPassword)
	if err != nil {
		log.Error("load client certificate fail: ", err)
	} else {
		s.transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: clientCrt.TlsCertificates(),
			},
		}

		caCrt := &gcrt.Crt{}
		err = caCrt.FromFile(caCrtFilePath)
		if err != nil {
			log.Error("load ca certificate fail: ", err)
		} else {
			s.transport.TLSClientConfig.RootCAs = caCrt.Pool()
		}
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runTestApi()
		s.runTestDoc()
	}()

	ch := make(chan error)
	go func(err chan<- error) {
		time.Sleep(time.Second)
		s.wg.Wait()
		err <- nil
	}(ch)

	return ch
}

func (s *Client) runTestApi() {
	// http
	httpClient := gclient.Http{}
	url := s.getTestApiUrl("http", uriHello)
	_, output, connState, _, err := httpClient.PostJson(url, nil)
	if err != nil {
		log.Error("test api hello fail: ", err)
	} else {
		if connState != nil {
			log.Error("connState: ", connState)
		}
		result := &gtype.Result{}
		err = result.Unmarshal(output)
		if err != nil {
			log.Error("test api hello fail: ", err)
		} else {
			log.Info("test api hello success")
			fmt.Println(result.FormatString())
			if result.Code != 0 {
				log.Error("test api hello fail: error code = ", result.Code)
			}
		}
	}

	// https
	url = s.getTestApiUrl("https", uriHello)
	_, output, connState, _, err = httpClient.PostJson(url, nil)
	if err == nil {
		log.Error("test api hello fail: ", "no transport should be err")
	} else {
		log.Debug("no transport error info: ", err)
	}

	httpClient.Transport = s.transport
	_, output, connState, _, err = httpClient.PostJson(url, nil)
	if err != nil {
		log.Error("test api hello fail: ", err)
	} else {
		if connState == nil {
			log.Error("connState: nil")
		} else {
			serverCrt := &gcrt.Crt{}
			serverCrt.FromConnectionState(connState)
			sou := serverCrt.OrganizationalUnit()
			if sou != serverCrtOU {
				log.Error("test api hello server organization unit invalid: ", sou)
			}
		}
		result := &gtype.Result{}
		err = result.Unmarshal(output)
		if err != nil {
			log.Error("test api hello fail: ", err)
		} else {
			log.Info("test api hello success")
			fmt.Println(result.FormatString())
			if result.Code != 0 {
				log.Error("test api hello fail: error code = ", result.Code)
			}
		}
	}
}

func (s *Client) runTestDoc() {

}

func (s *Client) getTestApiUrl(schema, uri string) string {
	port := cfg.Http.Port
	if schema == "https" {
		port = cfg.Https.Port
	}
	return fmt.Sprintf("%s://localhost:%d%s", schema, port, path.Uri(uri).Path())
}
