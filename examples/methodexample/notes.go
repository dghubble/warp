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
	mux.GetFunc("/notes/new", newHandler)
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
