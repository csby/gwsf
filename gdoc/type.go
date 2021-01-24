package gdoc

type Type struct {
	index  int
	Name   string   `json:"name"`   // 名称
	Fields []*Model `json:"fields"` // 成员
}

type TypeCollection []*Type

func (s TypeCollection) Len() int {
	return len(s)
}

func (s TypeCollection) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s TypeCollection) Less(i, j int) bool {
	return s[i].index < s[j].index
}
