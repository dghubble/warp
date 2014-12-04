package warp

import (
	"net/http"
)

// Route is an entry in a ServeMux routes map. It pairs a pattern with a
// handler and a slice of rules that the request should pass.
type Route struct {
	handler  http.Handler // handler for the route
	pattern  string       // pattern to report that the request matched
	implicit bool         // true for implicit routes added by ServeMux
	rules    []Rule       // route Rules
}

// NewRoute allocates and returns a new *Route.
func NewRoute(pattern string, handler http.Handler, rules ...Rule) *Route {
	return &Route{
		pattern: pattern,
		handler: handler,
		rules:   rules,
	}
}

// Allows returns true if each of its Rules Allows the request.
func (route *Route) Allows(request *http.Request) bool {
	for _, rule := range route.rules {
		if !rule.Allows(request) {
			return false
		}
	}
	return true
}

// Methods adds a MethodRule to the Route to constrain it to
// the specified methods:
//
//	mux := warp.NewServeMux()
//	mux.Register("/get-or-post", myHandler).Methods("GET", "POST")
func (route *Route) Methods(methods ...string) *Route {
	route.rules = append(route.rules, NewMethodRule(methods...))
	return route
}
