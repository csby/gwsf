package gdoc

import "encoding/json"

type MapData struct {
	kv map[string]interface{}
}

func (s *MapData) SetValue(k string, v interface{}) {
	if s.kv == nil {
		s.kv = make(map[string]interface{})
	}

	s.kv[k] = v
}

func (s MapData) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.kv)
}
