package main

import (
	"fmt"
	"github.com/dghubble/warp"
	"log"
	"net/http"
)

var mux *warp.ServeMux = warp.NewServeMux()

func init() {
	mux.Get("/notes", http.HandlerFunc(listHandler))
	mux.Get("/notes/new", http.HandlerFunc(newHandler))
	mux.Post("/notes", http.HandlerFunc(createHandler))
	mux.Get("/notes/:id", http.HandlerFunc(readHandler))
	mux.Put("/notes/:id", http.HandlerFunc(updateHandler))
	mux.Delete("/notes/:id", http.HandlerFunc(deleteHandler))
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

func newHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "new")
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
