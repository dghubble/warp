package warp

import (
	"net/http"
	"testing"
)

func BenchmarkRouteMatching(b *testing.B) {
	mux := NewServeMux()
	handler := func(w http.ResponseWriter, r *http.Request) {}
	mux.Get("/foo/bar", http.HandlerFunc(handler))
	b.ReportAllocs()
	req, _ := http.NewRequest("GET", "/foo/bar", nil)
	for i := 0; i < b.N; i++ {
		mux.ServeHTTP(nil, req)
	}
}
