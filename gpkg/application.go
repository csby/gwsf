package gpkg

type Application struct {
	Enable bool      `json:"enable"`
	Name   string    `json:"name"`
	Bin    Binary    `json:"bin"`
	Src    Source    `json:"src"`
	Webs   []Website `json:"webs"`
}
