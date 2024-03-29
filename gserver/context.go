package gserver

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/csby/gwsf/gtype"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type context struct {
	response    http.ResponseWriter
	request     *http.Request
	method      string
	schema      string
	host        string
	path        string
	token       string
	node        string
	instance    string
	forwardFrom string
	handled     bool

	certificate  gtype.Certificate
	queries      gtype.QueryCollection
	input        []byte
	inputFormat  int
	output       []byte
	outputFormat int
	outputCode   *int
	log          bool
	rid          uint64
	rip          string
	result       int
	enterTime    time.Time
	leaveTime    *time.Time
	keys         map[string]interface{}

	clientOrganizationalUnit string

	afterInput func(ctx *context)
}

func (s *context) Request() *http.Request {
	return s.request
}

func (s *context) Response() http.ResponseWriter {
	return s.response
}

func (s *context) Query(name string) string {
	if s.queries == nil {
		return ""
	}

	return s.queries.Value(name)
}

func (s *context) GetBody() ([]byte, error) {
	return ioutil.ReadAll(s.request.Body)
}

func (s *context) GetJson(v interface{}) error {
	err := json.NewDecoder(s.request.Body).Decode(v)
	if err == nil {
		s.input, err = json.Marshal(v)
	}
	s.inputFormat = gtype.ArgsFmtJson

	if s.afterInput != nil {
		go s.afterInput(s)
	}

	return err
}

func (s *context) GetXml(v interface{}) error {
	bodyData, err := ioutil.ReadAll(s.request.Body)
	if err != nil {
		return err
	}
	defer s.request.Body.Close()
	s.input = bodyData
	s.inputFormat = gtype.ArgsFmtXml

	err = xml.Unmarshal(bodyData, v)

	if s.afterInput != nil {
		go s.afterInput(s)
	}

	return err
}

func (s *context) GetSoapAction() string {
	if s.request == nil {
		return ""
	}

	if len(s.request.Header) < 1 {
		return ""
	}
	for k, v := range s.request.Header {
		if strings.ToLower(k) != "content-type" {
			continue
		}
		for i := 0; i < len(v); i++ {
			content := v[i]
			if len(content) < 1 {
				continue
			}
			items := strings.Split(content, ";")
			for j := 0; j < len(items); j++ {
				item := items[j]
				if len(item) < 1 {
					continue
				}
				subs := strings.Split(item, "=")
				if len(subs) == 2 {
					if strings.ToLower(subs[0]) == "action" {
						return strings.ReplaceAll(subs[1], "\"", "")
					}
				}
			}
		}

		break
	}

	return ""
}

func (s *context) OutputJson(v interface{}) {
	s.outputFormat = gtype.ArgsFmtJson
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Fprint(s.response, err)
	} else {
		s.response.Header().Add("Access-Control-Allow-Origin", "*")
		s.response.Header().Set("Content-Type", "application/json;charset=utf-8")
		s.response.Write(data)
		s.output = data
	}

}

func (s *context) OutputXml(v interface{}) {
	s.outputFormat = gtype.ArgsFmtXml
	if s.response == nil {
		return
	}

	if v != nil {
		switch v.(type) {
		case []byte:
			s.output = v.([]byte)
		case string:
			s.output = []byte(v.(string))
		default:
			bodyData, err := xml.MarshalIndent(v, "", "	")
			if err != nil {
				fmt.Fprint(s.response, err)
				s.output = []byte(err.Error())
				return
			} else {
				s.output = bodyData
			}
		}
	}

	s.response.Header().Add("Access-Control-Allow-Origin", "*")
	s.response.Header().Set("Content-Type", "application/xml;charset=utf-8")
	if len(s.output) > 0 {
		s.response.Write(s.output)
	}
}

func (s *context) OutputSoap(v interface{}) {
	s.outputFormat = gtype.ArgsFmtXml
	if s.response == nil {
		return
	}

	if v != nil {
		switch v.(type) {
		case []byte:
			s.output = v.([]byte)
		case string:
			s.output = []byte(v.(string))
		default:
			bodyData, err := xml.MarshalIndent(v, "", "	")
			if err != nil {
				fmt.Fprint(s.response, err)
				s.output = []byte(err.Error())
				return
			} else {
				s.output = bodyData
			}
		}
	}

	s.response.Header().Add("Access-Control-Allow-Origin", "*")
	s.response.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	if len(s.output) > 0 {
		s.response.Write(s.output)
	}
}

func (s *context) Success(data interface{}) {
	result := &gtype.Result{
		Code:   0,
		Data:   data,
		Elapse: time.Now().Sub(s.enterTime).String(),
		Serial: s.rid,
	}
	s.outputCode = &result.Code

	s.OutputJson(result)
}

func (s *context) Error(err gtype.Error, detail ...interface{}) {
	result := &gtype.Result{
		Code:   err.Code(),
		Elapse: time.Now().Sub(s.enterTime).String(),
		Serial: s.rid,
	}
	result.Error.Summary = err.Summary()
	details := s.toString(detail)
	if len(details) > 0 {
		result.Error.Detail = details
	} else {
		result.Error.Detail = err.Detail()
	}

	s.outputCode = &result.Code

	s.OutputJson(result)
}

func (s *context) ErrorWithData(data interface{}, err gtype.Error, detail ...interface{}) {
	result := &gtype.Result{
		Code:   err.Code(),
		Elapse: time.Now().Sub(s.enterTime).String(),
		Serial: s.rid,
		Data:   data,
	}
	result.Error.Summary = err.Summary()
	details := s.toString(detail)
	if len(details) > 0 {
		result.Error.Detail = details
	} else {
		result.Error.Detail = err.Detail()
	}

	s.outputCode = &result.Code

	s.OutputJson(result)
}

func (s *context) IsError() bool {
	if s.outputCode == nil {
		return false
	}

	if *s.outputCode == 0 {
		return false
	}

	return true
}

func (s *context) Method() string {
	return s.method
}

func (s *context) Schema() string {
	return s.schema
}

func (s *context) Host() string {
	return s.host
}

func (s *context) Path() string {
	return s.path
}

func (s *context) Queries() gtype.QueryCollection {
	return s.queries
}

func (s *context) Certificate() *gtype.Certificate {
	return &s.certificate
}

func (s *context) SetHandled(v bool) {
	s.handled = v
}

func (s *context) IsHandled() bool {
	return s.handled
}

func (s *context) Token() string {
	return s.token
}

func (s *context) Node() string {
	return s.node
}

func (s *context) Instance() string {
	return s.instance
}

func (s *context) ForwardFrom() string {
	return s.forwardFrom
}

func (s *context) GetQuery() []byte {
	if s.queries != nil {
		query, err := json.Marshal(s.queries)
		if err == nil {
			return query
		}
	}

	return nil
}

func (s *context) SetInput(v []byte) {
	s.input = v
}

func (s *context) SetInputFormat(v int) {
	s.inputFormat = v
}

func (s *context) GetInput() []byte {
	return s.input
}

func (s *context) GetInputFormat() int {
	return s.inputFormat
}

func (s *context) GetOutput() []byte {
	return s.output
}

func (s *context) GetOutputFormat() int {
	return s.outputFormat
}

func (s *context) GetOutputCode() *int {
	return s.outputCode
}

func (s *context) EnterTime() time.Time {
	return s.enterTime
}

func (s *context) LeaveTime() *time.Time {
	return s.leaveTime
}

func (s *context) RID() uint64 {
	return s.rid
}

func (s *context) RIP() string {
	return s.rip
}

func (s *context) SetLog(v bool) {
	s.log = v
}

func (s *context) GetLog() bool {
	return s.log
}

func (s *context) Set(key string, val interface{}) {
	s.keys[key] = val
}

func (s *context) Get(key string) (interface{}, bool) {
	val, ok := s.keys[key]
	if ok {
		return val, true
	} else {
		return nil, false
	}
}

func (s *context) Del(key string) bool {
	_, ok := s.keys[key]
	if ok {
		delete(s.keys, key)
		return true
	} else {
		return false
	}
}

func (s *context) ClientOrganization() string {
	return s.clientOrganizationalUnit
}

func (s *context) SetClientOrganization(ou string) {
	s.clientOrganizationalUnit = ou
}

func (s *context) NewGuid() string {
	return gtype.NewGuid()
}

func (s *context) toString(a []interface{}) string {
	sb := &strings.Builder{}
	c := len(a)
	for i := 0; i < c; i++ {
		item := a[i]
		if item == nil {
			continue
		}
		sb.WriteString(fmt.Sprint(item))
	}

	return sb.String()
}

func (s *context) FireAfterInput() {
	if s.afterInput == nil {
		return
	}

	go s.afterInput(s)
}
