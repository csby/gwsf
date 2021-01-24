package gdoc

type Error struct {
	Code    int    `json:"code" note:"错误代码"`
	Summary string `json:"summary" note:"错误描述"`
}

type ErrorCollection []*Error

func (s ErrorCollection) Len() int {
	return len(s)
}

func (s ErrorCollection) Less(i, j int) bool {
	if s[i].Code < s[j].Code {
		return true
	}

	return false
}

func (s ErrorCollection) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
