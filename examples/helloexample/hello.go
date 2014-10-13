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

func 你好处理(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hello, %s!\n", req.URL.Query().Get(":名"))
}
