package gtype

import "net/http"

type Router interface {
	GET(uri Uri, preHandle, httpHandle HttpHandle, docHandle DocHandle)
	POST(uri Uri, preHandle, httpHandle HttpHandle, docHandle DocHandle)

	// path of uri must be end with "/*filepath",
	// example: ServeFiles("/src/*filepath", http.Dir("/var/www"), nil)
	ServeFiles(uri Uri, preHandle HttpHandle, root http.FileSystem, docHandle DocHandle)

	// document
	Document() Doc
}
