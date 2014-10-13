# warp [![Build Status](https://travis-ci.org/dghubble/warp.png?branch=master)](https://travis-ci.org/dghubble/warp) [![GoDoc](http://godoc.org/github.com/dghubble/warp?status.png)](http://godoc.org/github.com/dghubble/warp)
 <img align="right" src="https://s3.amazonaws.com/dghubble/8-bit-gopher.png">

Package warp provides warp.ServeMux, an HTTP request multiplexer supporting routes with capture parameters, HTTP method requirements,
and other rule constriants. The mux matches incoming requests against
a list of registered routes and offers the following features:

* Routes can have capture params and matched parts of the URL can be
read from the query parameters. (e.g. `req.URL.Query().Get(":id")`).
* Routes can require requests to have particular HTTP Verb Methods.
* Routes can have additional matching rules based on the [http.Request](http://golang.org/pkg/net/http/#Request).
* Drop-in compatability with [http.ServeMux](http://golang.org/pkg/net/http/#ServeMux) 

The warp mux was originally forked from the standard [http.ServeMux](http://golang.org/pkg/net/http/#ServeMux) and
is compatible with it. The warp mux implements the same method 
signatures and passes all http.ServeMux tests (see [serve_test.go](serve_test.go)).

## Install

    $ go get github.com/dghubble/warp

## Usage

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
      mux.GetFunc("/你好/:名", 你好处理) // GET only
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
collection of rules that must be satisfied for the request to match the 
route. HTTP Method (GET, POST, etc.) rule requirements are quite common
so ServeMux provides convenience methods `mux.Get(pattern, handler)`, 
etc. for each verb. Convenience methods `mux.GetFunc(pattern, handlerFunc)`, etc. are also provided for each verb to accept handler functions.

    func init() {
      mux.GetFunc("/notes", listHandler)
      mux.GetFunc("/notes/new", newHandler)
      mux.PostFunc("/notes", createHandler)
      mux.GetFunc("/notes/:id", readHandler)
      mux.PutFunc("/notes/:id", updateHandler)
      mux.DelFunc("/notes/:id", deleteHandler)
    }

To register ServeMux routes with rules directly, use `HandleRoute`
or `HandleRouteFunc`.

## Full Docs

[https://godoc.org/github.com/dghubble/warp](https://godoc.org/github.com/dghubble/warp)

## Performance

    $ go test -bench .
    PASS
    BenchmarkRouteMatching   1000000        2486 ns/op

## License

[BSD](License)

## Thanks

Warp's ServeMux combines the high quality http.ServeMux base,
the pattern parameter ideas of [bmizerany/pat](https://github.com/bmizerany/pat), and the rule based matching ideas of [gorilla/mux](https://github.com/gorilla/mux). Thank you to each of these projects!
