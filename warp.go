package warp

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
)

// ServeMux is an HTTP request multiplexer.
// It matches the URL of each incoming request against a list of registered
// patterns and calls the handler for the pattern that
// most closely matches the URL.
//
// Patterns name fixed, rooted paths, like "/favicon.ico",
// or rooted subtrees, like "/images/" (note the trailing slash).
// Longer patterns take precedence over shorter ones, so that
// if there are handlers registered for both "/images/"
// and "/images/thumbnails/", the latter handler will be
// called for paths beginning "/images/thumbnails/" and the
// former will receive requests for any other paths in the
// "/images/" subtree.
//
// Note that since a pattern ending in a slash names a rooted subtree,
// the pattern "/" matches all paths not matched by other registered
// patterns, not just the URL with Path == "/".
//
// Patterns may optionally begin with a host name, restricting matches to
// URLs on that host only.  Host-specific patterns take precedence over
// general patterns, so that a handler might register for the two patterns
// "/codesearch" and "codesearch.google.com/" without also taking over
// requests for "http://www.google.com/".
//
// ServeMux also takes care of sanitizing the URL request path,
// redirecting any request containing . or .. elements to an
// equivalent .- and ..-free URL.
type ServeMux struct {
	mu       sync.RWMutex
	routes   map[string][]*Route // pattern -> routes
	anyHosts bool                // whether any patterns contain hostnames
}

// NewServeMux allocates and returns a new *ServeMux.
func NewServeMux() *ServeMux {
	return &ServeMux{
		routes: make(map[string][]*Route),
	}
}

// Handle registers the handler for the given pattern. Handle panics if the
// pattern is empty or the handler is nil.
func (mux *ServeMux) Handle(pattern string, handler http.Handler) {
	mux.addRoute(pattern, NewRoute(pattern, handler))
}

// HandleFunc registers the handler function for the given pattern.
func (mux *ServeMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	mux.Handle(pattern, http.HandlerFunc(handler))
}

// HandleRoute registers the handler for the pattern and rules. HandleRoute
// returns the new Route entry.
func (mux *ServeMux) HandleRoute(pattern string, handler http.Handler, rules ...rule) *Route {
	route := NewRoute(pattern, handler, rules...)
	mux.addRoute(pattern, route)
	return route
}

// HandleRouteFunc registers the handler function for the pattern and rules.
// Returns the new Route entry.
func (mux *ServeMux) HandleRouteFunc(pattern string, handler func(http.ResponseWriter, *http.Request), rules ...rule) *Route {
	return mux.HandleRoute(pattern, http.HandlerFunc(handler), rules...)
}

// Get registers the handler for the pattern and GET requests only. Returns
// the new Route entry.
func (mux *ServeMux) Get(pattern string, handler http.Handler) *Route {
	return mux.HandleRoute(pattern, handler, NewMethodRule("GET"))
}

// Post registers the handler for the pattern and POST requests only. Returns
// the new Route entry.
func (mux *ServeMux) Post(pattern string, handler http.Handler) *Route {
	return mux.HandleRoute(pattern, handler, NewMethodRule("POST"))
}

// Put registers the handler for the pattern and PUT requests only. Returns
// the new Route entry.
func (mux *ServeMux) Put(pattern string, handler http.Handler) *Route {
	return mux.HandleRoute(pattern, handler, NewMethodRule("PUT"))
}

// Delete registers the handler for the pattern and DELETE requests only.
// Returns the new Route entry.
func (mux *ServeMux) Del(pattern string, handler http.Handler) *Route {
	return mux.HandleRoute(pattern, handler, NewMethodRule("DELETE"))
}

// GetFunc registers the handler function for the pattern and GET requests
// only. Returns the new Route entry.
func (mux *ServeMux) GetFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) *Route {
	return mux.Get(pattern, http.HandlerFunc(handler))
}

// PostFunc registers the handler function for the pattern and POST requests
// only. Returns the new Route entry.
func (mux *ServeMux) PostFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) *Route {
	return mux.Post(pattern, http.HandlerFunc(handler))
}

// PutFunc registers the handler function for the pattern and PUT requests
// only. Returns the new Route entry.
func (mux *ServeMux) PutFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) *Route {
	return mux.Put(pattern, http.HandlerFunc(handler))
}

// DeleteFunc registers the handler function for the pattern and DELETE
// requests only. Returns the new Route entry.
func (mux *ServeMux) DelFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) *Route {
	return mux.Del(pattern, http.HandlerFunc(handler))
}

// Handler returns the handler to use for the given request,
// consulting r.Method, r.Host, and r.URL.Path. It always returns
// a non-nil handler. If the path is not in its canonical form, the
// handler will be an internally-generated handler that redirects
// to the canonical path.
//
// Handler also returns the registered pattern that matches the
// request or, in the case of internally-generated redirects,
// the pattern that will match after following the redirect.
//
// If there is no registered handler that applies to the request,
// Handler returns a ``page not found'' handler and an empty pattern.
func (mux *ServeMux) Handler(request *http.Request) (h http.Handler, pattern string) {
	if request.Method != "CONNECT" {
		if cleanedPath := cleanPath(request.URL.Path); cleanedPath != request.URL.Path {
			url := *request.URL
			url.Path = cleanedPath
			_, pattern = mux.handler(request, cleanedPath)
			return http.RedirectHandler(url.String(), http.StatusMovedPermanently), pattern
		}
	}
	return mux.handler(request, request.URL.Path)
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	handler, pattern := mux.Handler(r)
	// add capture params to query params
	_, _, params := pathMatch(pattern, r.URL.Path)
	if len(params) > 0 {
		r.URL.RawQuery = url.Values(params).Encode() + "&" + r.URL.RawQuery
	}
	handler.ServeHTTP(w, r)
}

// String returns a string representation of the registered routes
func (mux *ServeMux) String() string {
	mux.mu.RLock()
	defer mux.mu.RUnlock()
	var buf bytes.Buffer
	for pattern, _ := range mux.routes {
		for _, route := range mux.routes[pattern] {
			fmt.Fprintf(&buf, "%s -> %+v \n", pattern, route)
		}
	}
	return buf.String()
}

// addRoute registers the pattern for the handler for requests with the given
// HTTP method. If the pattern is a /tree/, inserts an implicit permanent
// redirect for /tree to /tree/ (provided no implicit /tree route exists). If
// the pattern is empty or the handler is nil, add panics.
func (mux *ServeMux) addRoute(pattern string, route *Route) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if route.pattern == "" {
		panic("warp: invalid pattern " + pattern)
	}
	if route.handler == nil {
		panic("warp: nil handler")
	}

	mux.routes[pattern] = append(mux.routes[pattern], route)

	// if registering the first pattern with a hostname
	if !mux.anyHosts && len(pattern) > 0 && pattern[0] != '/' {
		mux.anyHosts = true
	}

	// check if pattern is a /tree/ and no implicit route exists for the pattern
	// and insert a /tree -> /tree/ permanent redirect. If the pattern contains
	// a hostname, it is stripped from the redirection target url. Note that the
	// pattern key is /tree, but the route pattern is /tree/ for compliance with
	// http.ServeMux.Handler behavior and tests.
	n := len(pattern)
	if n > 1 && pattern[n-1] == '/' && !mux.hasImplicitRoute(pattern[:n-1]) {
		target := pattern
		// if pattern has a hostname, strip it from the target
		if pattern[0] != '/' {
			target = pattern[strings.Index(pattern, "/"):]
		}
		route := &Route{pattern, http.RedirectHandler(target, http.StatusMovedPermanently), true, nil}
		mux.routes[pattern[:n-1]] = append(mux.routes[pattern[:n-1]], route)
	}
}

// hasImplicitRoute returns true if the pattern has an implicit route (i.e.
// added by ServeMux), false otherwise.
func (mux *ServeMux) hasImplicitRoute(pattern string) bool {
	for _, route := range mux.routes[pattern] {
		if route.implicit {
			return true
		}
	}
	return false
}

// handler is the main implementation of Handler. Matches the path to a route
// and returns the handler and pattern of the route. Host-specific patterns
// take precedence over generic patterns. The given path is assumed to be the
// canonical (cleaned) request.URL.Path, except for CONNECT methods. Returns a
// NotFoundHandler and empty string pattern if no route matches.
func (mux *ServeMux) handler(request *http.Request, path string) (handler http.Handler, pattern string) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	// host-specific patterns
	if mux.anyHosts {
		handler, pattern, _ = mux.match(request, request.Host+path)
	}
	// generic patterns
	if handler == nil {
		handler, pattern, _ = mux.match(request, path)
	}
	// no handler found
	if handler == nil {
		handler, pattern = http.NotFoundHandler(), ""
	}
	return handler, pattern
}

// match will find the route that most closely matches the request. It first
// checks the request path against registered patterns for different route
// sets. Then, for routes matching the pattern, it checks that the request
// matches the route rules. In decreasing importance, longer patterns (more
// specific), explicit routes, and more capture params are preferred.
// Examples:
// Path /foo/bar/ matches /foo/bar/ over /foo/
// Path /explicit matches registered /explicit route over an implicit /explicit
// -> /explicit/ redirect from registering /explicit/
// Path /notes/new matches /notes/new over /notes/:id
// Path /site/i matches /site/:name over /site/
func (mux *ServeMux) match(request *http.Request, path string) (handler http.Handler, reportPattern string, params url.Values) {
	var n = 0 // num runes matched in previous best match
	for pattern, routes := range mux.routes {
		// skip patterns that the path doesn't match
		isMatch, runeCount, parameters := pathMatch(pattern, path)
		if !isMatch {
			continue
		}
		for _, route := range routes {
			// skip routes with rules that don't allow the request
			if !route.allows(request) {
				continue
			}
			// prefer longer patterns
			if handler == nil || runeCount > n {
				n = runeCount
				handler = route.handler
				// redirect route's pattern differs from pattern key
				reportPattern = route.pattern
				params = parameters
			}

			if runeCount == n {
				// prefer explicit routes, longer patterns excluding param names
				if !route.implicit {
					handler = route.handler
					reportPattern = route.pattern
					params = parameters
				}
			}
		}
	}
	return handler, reportPattern, params
}

// pathMatch returns whether the path matches the given pattern, how many
// runes matched, and the map of parameters captured from the path. /leaf
// patterns require the path to match exactly, while /tree/ patterns only
// require the path to start with /tree/ (so pattern / matches all paths).
func pathMatch(pattern, path string) (bool, int, url.Values) {
	var params = make(url.Values)
	var runeCount = 0

	if len(pattern) == 0 {
		// should not happen
		return false, runeCount, nil
	}

	// if pattern equals path, the path matches and the pattern has no capture params
	if pattern == path {
		return true, len([]rune(pattern)), nil
	}

	rPattern := []rune(pattern)
	rPath := []rune(path)
	n := len(rPattern)
	m := len(rPath)
	var i, j int
	// traverse pattern runes, capture params, compare to path runes
	for i < n {
		switch {
		case j >= m: // reached path end, but pattern has more runes
			return false, runeCount, nil
		case rPattern[i] == ':':
			var name, value string
			var next rune
			name, i, next = captureName(rPattern, i+1) // param name after ':'
			value, j = captureValue(rPath, j, next)
			params.Add(":"+name, value)
		case rPattern[i] == rPath[j]:
			i++
			j++
			runeCount++
		default:
			return false, runeCount, nil
		}
	}

	// if pattern is a /tree/, path need only start with the pattern
	if rPattern[n-1] == '/' {
		return true, runeCount, params
	}
	// otherwise, /leaf pattern so path indexes 0 through len(path) should
	// have matched the pattern
	if j != m {
		return false, runeCount, nil
	}
	return true, runeCount, params
}

// captureName captures the param name starting at the given rune index from
// the pattern. Returns the captured name, the next rune index, and the next
// non-variable rune or the zero value rune if no runes remain.
func captureName(pattern []rune, i int) (string, int, rune) {
	var next rune // zero value rune
	var start = i
	// URL query params are encoded, so the :param names should be encoded
	// as well since some programs may assume all param names are escaped.
	for i < len(pattern) && isParamRune(pattern[i]) {
		i++
	}
	if i < len(pattern) {
		next = pattern[i]
	}
	return string(pattern[start:i]), i, next
}

// captureValue captures the param value starting at the given rune index
// in the path and not continuing past the given endRune. Returns the
// captured value and the next rune index after the captured value.
func captureValue(path []rune, j int, endMark rune) (string, int) {
	var start = j
	for j < len(path) && path[j] != endMark && path[j] != '/' {
		j++
	}
	return string(path[start:j]), j
}

// isUnescaped returns whether the rune is a reserved character that should
// be percent encoded. These runes are prohibited from pattern param names.
// https://en.wikipedia.org/wiki/Percent-encoding#Types_of_URI_characters
func isUnescaped(r rune) bool {
	switch r {
	case '!', '#', '$', '&', '\'', '(', ')', '*', '+', ',', '/', ':', ';',
		'=', '?', '@', '[', ']':
		return true
	default:
		return false
	}
}

// isParamRune returns true if the rune is allowed in a pattern :param name.
// Notably, '_' is allowed in names.
func isParamRune(r rune) bool {
	switch r {
	// pattern literals may reasonably be expected to continue at these runes
	case '%', '-', '.', '<', '>', '\\', '^', '`', '{', '|', '}', '~':
		return false
	default:
		// pattern :params may not contain unencoded characters
		return !isUnescaped(r)
	}
}

// cleanPath returns the canonical path, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}
