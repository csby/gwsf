package gclient

import "encoding/base64"

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (s *Header) SetBasicAuth(username, password string) {
	auth := username + ":" + password
	basic := base64.StdEncoding.EncodeToString([]byte(auth))

	s.Key = "Authorization"
	s.Value = "Basic " + basic
}
