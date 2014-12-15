package warp

import (
	"fmt"
	"net/http"
)

// verbs
const vGET = "GET"
const vPOST = "POST"
const vPUT = "PUT"
const vDELETE = "DELETE"
const vHEAD = "HEAD"
const vOPTIONS = "OPTIONS"
const vANY = "ANY"

var vALL []string = []string{vGET, vPOST, vPUT, vDELETE, vHEAD, vOPTIONS, vANY}

// Route is an entry in a ServeMux routes map. It pairs a pattern with a
// handler and a slice of rules that the request should pass.
type Route struct {
	pattern  string       // pattern to report that the request matched
	any      http.Handler // default handler
	get      http.Handler // GET handler
	post     http.Handler // POST handler
	put      http.Handler // PUT handler
	delete   http.Handler // DELETE handler
	head     http.Handler // HEAD handler
	options  http.Handler // OPTIONS handler
	implicit bool         // true for implicit routes added by ServeMux
}

// NewRoute allocates and returns a new *Route.
func NewRoute(pattern string, handler http.Handler, verb string) *Route {
	route := &Route{pattern: pattern}
	route.addHandler(verb, handler)
	return route
}

func (route *Route) addHandler(verb string, handler http.Handler) {
	if !contains(vALL, verb) {
		panic(fmt.Sprintf("Invalid route verb %s\n", verb))
	}
	switch verb {
	case vGET:
		route.get = handler
	case vPOST:
		route.post = handler
	case vPUT:
		route.put = handler
	case vDELETE:
		route.delete = handler
	case vHEAD:
		route.head = handler
	case vOPTIONS:
		route.options = handler
	case vANY:
		route.any = handler
	}
}

func (route *Route) getHandler(verb string) http.Handler {
	switch {
	case verb == vGET && route.get != nil:
		return route.get
	case verb == vPOST && route.post != nil:
		return route.post
	case verb == vPUT && route.put != nil:
		return route.put
	case verb == vDELETE && route.delete != nil:
		return route.delete
	case verb == vHEAD && route.head != nil:
		return route.head
	case verb == vOPTIONS && route.options != nil:
		return route.options
	case route.any != nil:
		return route.any
	}
	return nil
}
