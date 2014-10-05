package warp

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

// Handler and ServeHTTP tests, Method rules

var registerRoutes = []struct {
	pattern string
	message string // to create a handle that returns the string
	rules   []rule
}{
	{"/leaf", "leaf", nil},
	{"/tree/", "tree", nil},
	// registers route twice to test that implicit redirects not duplicated
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

	// // directory paths
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

func TestHandler(t *testing.T) {
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
			t.Errorf("%s %s -> code %d, want %d", ht.method, ht.url, w.Code, ht.code)
		}
		if pattern != ht.pattern {
			t.Errorf("%s %s -> pattern %s, want %s", ht.method, ht.url, pattern, ht.pattern)
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

// pattern capture parameters

var pathMatchTests = []struct {
	pattern   string
	path      string     // path to be matched
	isMatch   bool       // should the path match the pattern
	runeCount int        // expected number of matching, non-param runes
	params    url.Values // expected captured params key to value map
}{
	{"/", "/", true, 1, nil},
	{"/", "/another_url", true, 1, url.Values{}},

	{"/:name", "/tim", true, 1, url.Values{":name": {"tim"}}},
	{"/foo/:name", "/foo/tim", true, 5, url.Values{":name": {"tim"}}},
	// patterns without trailing slash require exact match, path cannot continue
	{"/foo/:name", "/foo/tim/", false, 5, nil},
	{"/foo/:name", "/foo/tim/extra", false, 5, nil},

	// pattern with trailing slash literal after param, path should have slash
	{"/foo/:name/", "/foo/tim", false, 5, nil},
	{"/foo/:name/", "/foo/tim/", true, 6, url.Values{":name": {"tim"}}},
	// patterns with trailing slash literal, path need only start with pattern
	{"/foo/:name/", "/foo/tim/extra", true, 6, url.Values{":name": {"tim"}}},

	// pattern with capture param between literal parts to be matched
	{"/foo/:name/bar", "/foo/bar", false, 5, nil},
	{"/foo/:name/bar", "/foo//bar", true, 9, url.Values{":name": {""}}},
	{"/foo/:name/bar", "/foo/tim/bar", true, 9, url.Values{":name": {"tim"}}},
	{"/foo/:name/bar", "/foo/tim/bar/", false, 9, nil},
	{"/foo/:name/bar", "/foo/tim/bar/extra", false, 9, nil},
	// slashes are never captured as part of a param's value
	{"/foo/:name/bar", "/foo/tim/extra/bar", false, 6, nil},
	{"/foo/:name.txt", "/foo/bar/tim.txt", false, 5, nil},

	{"/foo/:name/bar/", "/foo/tim/bar", false, 9, nil},
	{"/foo/:name/bar/", "/foo/tim/bar/", true, 10, url.Values{":name": {"tim"}}},
	{"/foo/:name/bar/", "/foo/tim/bar/extra", true, 10, url.Values{":name": {"tim"}}},
	{"/foo/:name/bar/", "/foo/tim/extra/bar/", false, 6, nil},

	// pattern with multiple capture params
	{"/foo/:name/bar/:id", "/foo/tim/bar", false, 9, nil},
	{"/foo/:name/bar/:id", "/foo/tim/bar/61", true, 10, url.Values{":name": {"tim"}, ":id": {"61"}}},
	{"/foo/:name/bar/:id", "/foo/tim/bar/61/", false, 10, nil},
	{"/foo/:name/bar/:id", "/foo/tim/bar/61/extra", false, 10, nil},

	// pattern with reuse of the same capture param
	{"/foo/:name/bar/:name", "/foo/ben/bar/tim", true, 10, url.Values{":name": {"ben", "tim"}}},
	// capture path value that uses a ':'
	{"/foo/:name", "/foo/:value", true, 5, url.Values{":name": {":value"}}},
	// dot in path is uncaptured
	{"/foo/:file.:ext", "/foo/cats.png", true, 6, url.Values{":file": {"cats"}, ":ext": {"png"}}},
	{"/foo/:file.:ext", "/foo/.png", true, 6, url.Values{":file": {""}, ":ext": {"png"}}},
	{"/foo/:name.txt", "/foo/tim.txt", true, 9, url.Values{":name": {"tim"}}},

	// pattern with capture param and literal at the same / level
	{"/foo/x:name", "/foo/tim", false, 5, nil},
	{"/foo/x:name", "/foo/xtim", true, 6, url.Values{":name": {"tim"}}},

	{"/안녕/:世界", "/안녕/tim", true, 4, url.Values{":世界": {"tim"}}},
	{"/안녕/:ם", "/안녕/世界", true, 4, url.Values{":ם": {"世界"}}},
}

func TestPathMatch(t *testing.T) {
	for _, pt := range pathMatchTests {
		isMatch, runeCount, params := pathMatch(pt.pattern, pt.path)
		if isMatch != pt.isMatch {
			t.Errorf("path %s match pattern %s, %t, want %t", pt.path, pt.pattern, isMatch, pt.isMatch)
		}
		if runeCount != pt.runeCount {
			t.Errorf("path %s match pattern %s, runeCount %d, want %d", pt.path, pt.pattern, runeCount, pt.runeCount)
		}
		if !reflect.DeepEqual(params, pt.params) {
			t.Errorf("path %s match pattern %s, params %v, want %v", pt.path, pt.pattern, params, pt.params)
		}
	}
}

var registerParamRoutes = []struct {
	pattern string
	rules   []rule
}{
	{"/foo/:name", nil},
	{"/bar/:name/", nil},
	{"/first/:age/last", nil},
	{"/begin/:start/end/:stop/", nil},
	{"/:reuse/:reuse", nil},
	{"github.com/:name", nil},
}

var emptyParams = make(url.Values)

var paramTests = []struct {
	method string     // test request method
	url    string     // test request url
	code   int        // expected HTTP response code
	params url.Values // expected captured params key to value map
}{
	// leaf paths
	{"GET", "/unknown", 404, emptyParams},
	{"GET", "/foo/tim", 200, url.Values{":name": {"tim"}}},
	{"GET", "/foo/tim/", 404, emptyParams},
	{"GET", "/bar/tim", 301, emptyParams},
	{"GET", "/bar/tim/", 200, url.Values{":name": {"tim"}}},
	{"GET", "/first/23/last", 200, url.Values{":age": {"23"}}},
	{"GET", "/first/23/last/", 404, emptyParams},
	{"GET", "/begin/0:00/end/16:54/", 200, url.Values{":start": {"0:00"}, ":stop": {"16:54"}}},
	// {"GET", "http://github.com/dghubble", 200, url.Values{":name": {"dghubble"}}},
}

func TestServeHTTPParams(t *testing.T) {
	mux := NewServeMux()
	for _, route := range registerParamRoutes {
		mux.HandleRoute(route.pattern, stringHandler("message"), route.rules...)
	}

	for _, pt := range paramTests {
		r := newRequest(pt.method, pt.url)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != pt.code {
			t.Errorf("path %s -> code %d, want %d", pt.url, w.Code, pt.code)
		}
		if !reflect.DeepEqual(r.URL.Query(), pt.params) {
			t.Errorf("path %s -> params %v, want %v", pt.url, r.URL.Query(), pt.params)
		}
	}
}
