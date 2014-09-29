package warp

import (
	"net/http"
)

// Route is an entry in a ServeMux routes map. It pairs a pattern with a
// handler and a slice of rules that the request should pass.
type Route struct {
	pattern  string       // pattern to report that the request matched
	handler  http.Handler // handler for the route entry
	implicit bool         // implicit route added by ServeMux
	rules    []rule       // rules HTTP method rules
}

// NewRoute allocates and returns a new *Route.
func NewRoute(pattern string, handler http.Handler, rules ...rule) *Route {
	return &Route{
		pattern:  pattern,
		handler:  handler,
		implicit: false,
		rules:    rules,
	}
}

// allows returns true if all route rules allow the request.
func (route *Route) allows(request *http.Request) bool {
	for _, rule := range route.rules {
		if !rule.allows(request) {
			return false
		}
	}
	return true
}
