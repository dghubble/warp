package warp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var registerRoutes = []struct {
	pattern string
	message string // to create a handle that returns the string
	rules   []rule
}{
	{"/leaf", "leaf", nil},
	{"/tree/", "tree", nil},
	// purposefully registers route twice to ensure implicit redirects not duplicated
	{"/tree/", "tree", nil},
	{"/explicit/", "explicit tree", nil},
	{"/explicit", "explicit leaf", nil},
	{"/get-only", "get only", []rule{NewMethodRule("GET")}},
	{"/post-or-put", "post or put only", []rule{NewMethodRule("POST", "PUT")}},
	{"/delete-only/", "delete only tree", []rule{NewMethodRule("DELETE")}},
	{"/delete-only", "allow post, override implicit redirect", []rule{NewMethodRule("POST")}},
}

var handlerTests = []struct {
	method  string // test request method
	url     string // test request url
	code    int    // expected HTTP response code
	pattern string // expected matching pattern
}{
	// leaf paths
	{"GET", "/leaf", 200, "/leaf"},
	// no trailing slash redirect for leaf patterns (no trailing slash)
	{"GET", "/leaf/", 404, ""},

	// unmatched path
	{"GET", "/unmatched", 404, ""},

	// directory paths
	{"GET", "/tree/", 200, "/tree/"},
	// ServeMux inserts implicit permanent redirect
	{"GET", "/tree", 301, "/tree/"},
	{"GET", "/explicit/", 200, "/explicit/"},
	// explicit route overrides implicit redirect added by ServeMux
	{"GET", "/explicit", 200, "/explicit"},

	// method-specific routes
	{"GET", "/get-only", 200, "/get-only"},
	{"POST", "/get-only", 404, ""},
	{"PUT", "/get-only", 404, ""},
	{"DELETE", "/get-only", 404, ""},
	{"GET", "/post-or-put", 404, ""},
	{"POST", "/post-or-put", 200, "/post-or-put"},
	{"PUT", "/post-or-put", 200, "/post-or-put"},
	{"DELETE", "/post-or-put", 404, ""},
	{"GET", "/delete-only/", 404, ""},
	{"POST", "/delete-only/", 404, ""},
	{"PUT", "/delete-only/", 404, ""},
	{"DELETE", "/delete-only/", 200, "/delete-only/"},
	// implicit redirects are not method specific
	{"GET", "/delete-only", 301, "/delete-only/"},
	{"PUT", "/delete-only", 301, "/delete-only/"},
	{"DELETE", "/delete-only", 301, "/delete-only/"},
	// explicit route overrides implicit redirect added by ServeMux
	{"POST", "/delete-only", 200, "/delete-only"},
}

func newRequest(method, urlStr string) *http.Request {
	request, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		panic(err)
	}
	return request
}

func TestHandleRoute(t *testing.T) {
	mux := NewServeMux()
	for _, route := range registerRoutes {
		mux.HandleRoute(route.pattern, stringHandler(route.message), route.rules...)
	}

	for _, ht := range handlerTests {
		r := newRequest(ht.method, ht.url)
		handler, pattern := mux.Handler(r)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != ht.code {
			t.Errorf("%s %s -> code=%d, want %d", ht.method, ht.url, w.Code, ht.code)
		}
		if pattern != ht.pattern {
			t.Errorf("%s %s -> pattern=%s, want %s", ht.method, ht.url, pattern, ht.pattern)
		}
	}
}

var implicitRedirectPatterns = []string{"/tree", "/delete-only"}

// Tests that explicitly registering a /tree/ multiple times does not cause
// ServeMux to add duplicate useless /tree -> /tree/ implicit redirects.
func TestNoDuplicateImplicitRedirects(t *testing.T) {
	mux := NewServeMux()
	for _, route := range registerRoutes {
		mux.HandleRoute(route.pattern, stringHandler(route.message), route.rules...)
	}

	var count int
	for _, pattern := range implicitRedirectPatterns {
		count = 0
		for _, route := range mux.routes[pattern] {
			if route.implicit {
				count++
			}
		}
		if count > 1 {
			t.Errorf("pattern %s has %d implicit redirects, want 1", pattern, count)
		}
	}
}
