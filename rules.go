package warp

import (
	"net/http"
	"strings"
)

// Structs implementing the Rule interface can be used to constrain the
// requests a Route must handle.
// Allows returns true if the request passes the rule and may be handled
// or false if the request does not pass.
type Rule interface {
	Allows(*http.Request) bool
}

type methodRule []string

func NewMethodRule(methods ...string) methodRule {
	for i, method := range methods {
		methods[i] = strings.ToUpper(method)
	}
	return methodRule(methods)
}

// Allows returns true if the request.Method is in the allowed HTTP methods.
func (rule methodRule) Allows(request *http.Request) bool {
	return contains(rule, request.Method)
}
