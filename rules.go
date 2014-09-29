package warp

import (
	"net/http"
)

type rule interface {
	allows(*http.Request) bool
}

type methodRule []string

func NewMethodRule(methods ...string) methodRule {
	return methodRule(methods)
}

// allows returns true if the request.Method is in the allowed HTTP methods.
func (rule methodRule) allows(request *http.Request) bool {
	return contains(rule, request.Method)
}
