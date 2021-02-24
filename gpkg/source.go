package gpkg

type Source struct {
	Enable bool     `json:"enable"`
	Root   string   `json:"root"`
	Ignore []string `json:"ignore"`
}

func (s *Source) IsIgnore(name string) bool {
	if len(s.Ignore) < 1 {
		return false
	}

	for _, v := range s.Ignore {
		if v == name {
			return true
		}
	}

	return false
}
