package warp

import (
	"net/http"
	"testing"
)

func BenchmarkRouteMatching(b *testing.B) {
	mux := NewServeMux()
	handler := func(w http.ResponseWriter, r *http.Request) {}
	mux.GetFunc("/foo/:bar", handler)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		req, err := http.NewRequest("GET", "/foo/dghubble", nil)
		if err != nil {
			panic(err)
		}
		b.StartTimer()
		mux.ServeHTTP(nil, req)
	}
}
