package gpkg

type Binary struct {
	Root  string            `json:"root"`
	Files map[string]string `json:"files"`
}
