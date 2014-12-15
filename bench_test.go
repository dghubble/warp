package warp

import (
	"net/http"
	"testing"
)

func BenchmarkRouteMatching(b *testing.B) {
	mux := NewServeMux()
	handler := func(w http.ResponseWriter, r *http.Request) {}
	mux.Get("/foo/:bar", http.HandlerFunc(handler))
	req, err := http.NewRequest("GET", "/foo/dghubble", nil)
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		mux.ServeHTTP(nil, req)
	}
}
