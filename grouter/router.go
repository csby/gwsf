package grouter

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"github.com/csby/gwsf/gtype"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Handle is a function that can be registered to a route to handle HTTP
// requests. Like http.HandlerFunc, but has a third parameter for the values of
// wildcards (path variables).
type Handle func(http.ResponseWriter, *http.Request, gtype.Params)

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	trees map[string]*node

	paramsPool sync.Pool
	maxParams  uint16

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 308 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 308 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// If enabled, the router automatically replies to OPTIONS requests.
	// Custom OPTIONS handlers take priority over automatic replies.
	HandleOPTIONS bool

	// An optional http.Handler that is called on automatic OPTIONS requests.
	// The handler is only called if HandleOPTIONS is true and no OPTIONS
	// handler for the specific path was set.
	// The "Allowed" header is set before calling the handler.
	GlobalOPTIONS http.Handler

	// Cached value of global (*) allowed methods
	globalAllowed string

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	NotFound http.Handler

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed http.Handler

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})

	// Document
	Doc gtype.Doc
}

// New returns a new initialized Router.
// Path auto-correction, including trailing slashes, is enabled by default.
func New() *Router {
	return &Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}
}

func (s *Router) Document() gtype.Doc {
	return s.Doc
}

func (s *Router) getParams() *gtype.Params {
	ps := s.paramsPool.Get().(*gtype.Params)
	*ps = (*ps)[0:0] // reset slice
	return ps
}

func (s *Router) putParams(ps *gtype.Params) {
	if ps != nil {
		s.paramsPool.Put(ps)
	}
}

// GET is a shortcut for router.Handle(http.MethodGet, path, handle)
func (s *Router) GET(uri gtype.Uri, preHandle, httpHandle gtype.HttpHandle, docHandle gtype.DocHandle) {
	s.Handle(http.MethodGet, uri, preHandle, httpHandle, docHandle)
}

// HEAD is a shortcut for router.Handle(http.MethodHead, path, handle)
func (s *Router) HEAD(uri gtype.Uri, preHandle, httpHandle gtype.HttpHandle, docHandle gtype.DocHandle) {
	s.Handle(http.MethodHead, uri, preHandle, httpHandle, docHandle)
}

// OPTIONS is a shortcut for router.Handle(http.MethodOptions, path, handle)
func (s *Router) OPTIONS(uri gtype.Uri, preHandle, httpHandle gtype.HttpHandle, docHandle gtype.DocHandle) {
	s.Handle(http.MethodOptions, uri, preHandle, httpHandle, docHandle)
}

// POST is a shortcut for router.Handle(http.MethodPost, path, handle)
func (s *Router) POST(uri gtype.Uri, preHandle, httpHandle gtype.HttpHandle, docHandle gtype.DocHandle) {
	s.Handle(http.MethodPost, uri, preHandle, httpHandle, docHandle)
}

// PUT is a shortcut for router.Handle(http.MethodPut, path, handle)
func (s *Router) PUT(uri gtype.Uri, preHandle, httpHandle gtype.HttpHandle, docHandle gtype.DocHandle) {
	s.Handle(http.MethodPut, uri, preHandle, httpHandle, docHandle)
}

// PATCH is a shortcut for router.Handle(http.MethodPatch, path, handle)
func (s *Router) PATCH(uri gtype.Uri, preHandle, httpHandle gtype.HttpHandle, docHandle gtype.DocHandle) {
	s.Handle(http.MethodPatch, uri, preHandle, httpHandle, docHandle)
}

// DELETE is a shortcut for router.Handle(http.MethodDelete, path, handle)
func (s *Router) DELETE(uri gtype.Uri, preHandle, httpHandle gtype.HttpHandle, docHandle gtype.DocHandle) {
	s.Handle(http.MethodDelete, uri, preHandle, httpHandle, docHandle)
}

// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (s *Router) Handle(method string, uri gtype.Uri, before, handle gtype.HttpHandle, docHandle gtype.DocHandle) {
	varsCount := uint16(0)

	if method == "" {
		panic("method must not be empty")
	}
	path := uri.Path()
	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}
	if handle == nil {
		panic("handle must not be nil")
	}

	if s.trees == nil {
		s.trees = make(map[string]*node)
	}

	root := s.trees[method]
	if root == nil {
		root = new(node)
		s.trees[method] = root

		s.globalAllowed = s.allowed("*", "")
	}

	root.addRoute(path, handle, before)

	// Update maxParams
	if paramsCount := countParams(path); paramsCount+varsCount > s.maxParams {
		s.maxParams = paramsCount + varsCount
	}

	// Lazy-init paramsPool alloc func
	if s.paramsPool.New == nil && s.maxParams > 0 {
		s.paramsPool.New = func() interface{} {
			ps := make(gtype.Params, 0, s.maxParams)
			return &ps
		}
	}

	// document
	if docHandle != nil {
		if s.Doc != nil {
			if s.Doc.Enable() {
				docHandle(s.Doc, method, uri)
				s.Doc.Log(docHandle, method, uri)
			}
		}
	}
}

// ServeFiles serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use http.Dir:
//     router.ServeFiles("/src/*filepath", http.Dir("/var/www"))
func (s *Router) ServeFiles(uri gtype.Uri, preHandle gtype.HttpHandle, root http.FileSystem, docHandle gtype.DocHandle) {
	path := uri.Path()
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	s.GET(uri, preHandle, func(ctx gtype.Context, ps gtype.Params) {
		w := ctx.Response()
		req := ctx.Request()
		req.URL.Path = ps.ByName("filepath")

		acceptEncoding := req.Header.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
			filePath := filepath.Join(fmt.Sprint(root), req.URL.Path)
			contentType := mime.TypeByExtension(filepath.Ext(filePath))
			if len(contentType) > 0 {
				fileExisted := true
				fi, err := os.Stat(filePath)
				if err != nil {
					if os.IsNotExist(err) {
						fileExisted = false
					}
				} else {
					if fi.IsDir() {
						fileExisted = false
					}
				}
				if fileExisted {
					if s.serveFilesWithGZip(w, filePath, contentType) {
						return
					}
				}
			}
		}

		fileServer := http.FileServer(root)
		fileServer.ServeHTTP(w, req)
	}, docHandle)
}

func (s *Router) recv(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		s.PanicHandler(w, req, rcv)
	}
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (s *Router) Lookup(method, path string) (gtype.HttpHandle, gtype.HttpHandle, gtype.Params, bool) {
	if root := s.trees[method]; root != nil {
		handle, before, ps, tsr := root.getValue(path, s.getParams)
		if handle == nil {
			s.putParams(ps)
			return nil, nil, nil, tsr
		}
		if ps == nil {
			return handle, before, nil, tsr
		}
		return handle, before, *ps, tsr
	}
	return nil, nil, nil, false
}

func (s *Router) allowed(path, reqMethod string) (allow string) {
	allowed := make([]string, 0, 9)

	if path == "*" { // server-wide
		// empty method is used for internal calls to refresh the cache
		if reqMethod == "" {
			for method := range s.trees {
				if method == http.MethodOptions {
					continue
				}
				// Add request method to list of allowed methods
				allowed = append(allowed, method)
			}
		} else {
			return s.globalAllowed
		}
	} else { // specific path
		for method := range s.trees {
			// Skip the requested method - we already tried this one
			if method == reqMethod || method == http.MethodOptions {
				continue
			}

			handle, _, _, _ := s.trees[method].getValue(path, nil)
			if handle != nil {
				// Add request method to list of allowed methods
				allowed = append(allowed, method)
			}
		}
	}

	if len(allowed) > 0 {
		// Add request method to list of allowed methods
		allowed = append(allowed, http.MethodOptions)

		// Sort allowed methods.
		// sort.Strings(allowed) unfortunately causes unnecessary allocations
		// due to allowed being moved to the heap and interface conversion
		for i, l := 1, len(allowed); i < l; i++ {
			for j := i; j > 0 && allowed[j] < allowed[j-1]; j-- {
				allowed[j], allowed[j-1] = allowed[j-1], allowed[j]
			}
		}

		// return as comma separated list
		return strings.Join(allowed, ", ")
	}
	return
}

// ServeHTTP makes the router implement the http.Handler interface.
func (s *Router) Serve(ctx gtype.Context) {
	w := ctx.Response()
	req := ctx.Request()
	if s.PanicHandler != nil {
		defer s.recv(w, req)
	}

	path := req.URL.Path

	if root := s.trees[req.Method]; root != nil {
		if handle, before, ps, tsr := root.getValue(path, s.getParams); handle != nil {
			if ps != nil {
				if before != nil {
					before(ctx, *ps)
				}
				if !ctx.IsHandled() {
					handle(ctx, *ps)
				}
				s.putParams(ps)
			} else {
				if before != nil {
					before(ctx, nil)
				}
				if !ctx.IsHandled() {
					handle(ctx, nil)
				}
			}
			return
		} else if req.Method != http.MethodConnect && path != "/" {
			// Moved Permanently, request with GET method
			code := http.StatusMovedPermanently
			if req.Method != http.MethodGet {
				// Permanent Redirect, request with same method
				code = http.StatusPermanentRedirect
			}

			if tsr && s.RedirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					req.URL.Path = path[:len(path)-1]
				} else {
					req.URL.Path = path + "/"
				}
				http.Redirect(w, req, req.URL.String(), code)
				return
			}

			// Try to fix the request path
			if s.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					s.RedirectTrailingSlash,
				)
				if found {
					req.URL.Path = fixedPath
					http.Redirect(w, req, req.URL.String(), code)
					return
				}
			}
		}
	}

	if req.Method == http.MethodOptions && s.HandleOPTIONS {
		// Handle OPTIONS requests
		if allow := s.allowed(path, http.MethodOptions); allow != "" {
			w.Header().Set("Allow", allow)
			if s.GlobalOPTIONS != nil {
				s.GlobalOPTIONS.ServeHTTP(w, req)
			}
			return
		}
	} else if s.HandleMethodNotAllowed { // Handle 405
		if allow := s.allowed(path, req.Method); allow != "" {
			w.Header().Set("Allow", allow)
			if s.MethodNotAllowed != nil {
				s.MethodNotAllowed.ServeHTTP(w, req)
			} else {
				http.Error(w,
					http.StatusText(http.StatusMethodNotAllowed),
					http.StatusMethodNotAllowed,
				)
			}
			return
		}
	}

	// Handle 404
	if s.NotFound != nil {
		s.NotFound.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}

func (s *Router) serveFilesWithGZip(w http.ResponseWriter, filePath, contentType string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Add("Vary", "Accept-Encoding")

	gw, _ := gzip.NewWriterLevel(w, flate.BestCompression)
	defer gw.Close()

	io.Copy(gw, file)

	return true
}
