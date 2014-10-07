/*
Package warp provides a request router/mux with capture parameters,
HTTP verbs, and rules. Drop-in http.ServeMux compatability.

warp.ServeMux is an HTTP request multiplexer that matches incoming
requests against a list of registered routes and calls the handler
for the route that matches the route and its associated rules.

	* Routes can have capture params and matched parts of the URL can be
	read from the query parameters. (e.g. req.URL.Query().Get(":id")).
	* Routes can require requests to have particular HTTP Verb Methods.
	* Routes can have additional matching constraints based on the request.
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
		mux.HandleFunc("/hello/:name", helloHandler)
	}

	// main starts serving the web application
	func main() {
		address := "localhost:8080"
		log.Printf("Starting Server listening on %s\n", address)
		err := http.ListenAndServe(address, mux)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}

	// helloHandler writes a greeting
	func helloHandler(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello, %s!\n", req.URL.Query().Get(":name"))
	}

A Route struct collects together a pattern, its handler, and a
collection of rules that be satisfied for the request to match the
route. HTTP Method (GET, POST, etc.) rule requirements are quite common
so ServeMux provides convenience methods mux.Get(pattern, handler),
etc. for each verb. Convenience methods mux.GetFunc(pattern, handlerFunc),
etc. are also provided for each verb to accept handler functions.

	package main

	import (
		"fmt"
		"github.com/dghubble/warp"
		"log"
		"net/http"
	)

	var mux *warp.ServeMux = warp.NewServeMux()

	func init() {
		mux.GetFunc("/notes", listHandler)
		mux.PostFunc("/notes", createHandler)
		mux.GetFunc("/notes/:id", readHandler)
		mux.PutFunc("/notes/:id", updateHandler)
		mux.DelFunc("/notes/:id", deleteHandler)
	}

	// main starts serving the web application
	func main() {
		address := "localhost:8080"
		log.Printf("Starting Server listening on %s\n", address)
		err := http.ListenAndServe(address, mux)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}

	func listHandler(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "list")
	}

	func createHandler(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "create")
	}

	func readHandler(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "read %s", req.URL.Query().Get(":id"))
	}

	func updateHandler(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "update %s", req.URL.Query().Get(":id"))
	}

	func deleteHandler(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "delete %s", req.URL.Query().Get(":id"))
	}

To register ServeMux routes with rules directly, use HandleRoute
or HandleRouteFunc.
*/
package warp
