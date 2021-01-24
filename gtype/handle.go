package gtype

type (
	HttpHandle func(ctx Context, ps Params)
	DocHandle  func(doc Doc, method string, uri Uri)
)
