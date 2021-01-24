package gtype

import "strings"

type Query struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type QueryCollection []*Query

func (s QueryCollection) Value(key string) string {
	c := len(s)
	for i := 0; i < c; i++ {
		item := s[i]
		if strings.Compare(item.Key, key) == 0 {
			return item.Value
		}
	}

	return ""
}
