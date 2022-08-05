package gmodel

import (
	"encoding/xml"
	"io/ioutil"
)

type ServiceTomcatPom struct {
	Parent       ServiceTomcatPomParent `xml:"parent"`
	ModelVersion string                 `xml:"modelVersion"`
}

type ServiceTomcatPomParent struct {
	Version string `xml:"version"`
}

func (s *ServiceTomcatPom) LoadFromFile(path string) error {

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return xml.Unmarshal(bytes, s)
}
