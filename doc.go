/*
Package warp provides an HTTP ServeMux supporting rule constraints.

warp.ServeMux is an HTTP request multiplexer that matches incoming requests
against a list of registered routes and calls the handler for the route that
matches the route and its associated rules.

* Patterns can have rule constraints based on the request Method.
* Completely compatible with the standard http.ServeMux. warp.ServeMux
implements the same methods and passes all http.ServeMux tests.
*/
package warp
