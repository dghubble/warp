/*
Package warp provides a request router/mux with capture parameters,
HTTP verbs, and rules. Drop-in http.ServeMux compatability.

warp.ServeMux is an HTTP request multiplexer supporting routes with
capture parameters, HTTP method requirements, and other rule
constriants. The mux matches incoming requests against a list of
registered routes and offers the following features:

	* Routes can have capture params and matched parts of the URL can be
	read from the query parameters. (e.g. req.URL.Query().Get(":id")).
	* Routes can require requests to have particular HTTP Verb Methods.
	* Routes can have additional matching rules based on the request.
	* Drop-in compatability with http.ServeMux

The warp mux was originally forked from the standard http.ServeMux and
is compatible with it. The warp mux implements the same method
signatures and passes all http.ServeMux tests (see serve_test.go).

Handle and HandleFunc behave as they do in http.ServeMux, but allow
patterns with capture parameters for variable URL parts can be used
as well.

	package main

	import (
		"fmt"
		"github.com/dghubble/warp"
		"log"
		"net/http"
	)

	var mux *warp.ServeMux = warp.NewServeMux()

	func init() {
		mux.Handle("/hello/:name", http.HandlerFunc(helloHandler))
		mux.Get("/你好/:名", http.HandlerFunc(你好处理)) // GET only
	}

	// main starts serving the web application
	func main() {
		err := http.ListenAndServe("localhost:8080", mux)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}

	// helloHandler writes a greeting
	func helloHandler(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello, %s!\n", req.URL.Query().Get(":name"))
	}

	func 你好处理(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello, %s!\n", req.URL.Query().Get(":名"))
	}

A Route struct collects together a pattern, its handler, and a
collection of Rules that must be satisfied for the request to match the
route. HTTP Method (GET, POST, etc.) rule requirements are quite common
so ServeMux provides convenience methods mux.Get(pattern, handler),
etc. for each verb.

	func init() {
		mux.Get("/notes", http.HandlerFunc(listHandler))
		mux.Get("/notes/new", http.HandlerFunc(newHandler))
		mux.Post("/notes", http.HandlerFunc(createHandler))
		mux.Get("/notes/:id", http.HandlerFunc(readHandler))
		mux.Put("/notes/:id", http.HandlerFunc(updateHandler))
		mux.Delete("/notes/:id", http.HandlerFunc(deleteHandler))
	}

To register routes on a warp ServeMux directly, use the
`ServeMux.Register(pattern string, handler http.Handler, rules ...Rule) *Route`
method.
*/
package warp
