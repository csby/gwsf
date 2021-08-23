package gtype

type Page struct {
	PageIndex int64 `json:"pageIndex" note:"当前页码，从1开始，默认1"`
	PageSize  int64 `json:"pageSize" note:"页面大小(每页项目数)，默认15"`
}

type PageFilter struct {
	Page
	Filter interface{} `json:"filter" note:"过滤条件"`
}

type PageResult struct {
	Page
	PageCount int64       `json:"pageCount" note:"页数"`
	ItemCount int64       `json:"itemCount" note:"项目总数"`
	PageItems interface{} `json:"pageItems" note:"当前页项目列表"`
	Extend    interface{} `json:"extend,omitempty" note:"扩展信息"`
}
