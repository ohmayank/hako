package main

import (
	"log"
	"net/http"

	"github.com/ohmayank/hako/internal/handlers"
	"github.com/ohmayank/hako/internal/store"
)

func main() {
	s := store.NewFileStore("./.data")
	mux := http.NewServeMux()

	mux.HandleFunc("PUT /objects/{bucket}/{key...}", handlers.PutObject(s))
	mux.HandleFunc("GET /objects/{bucket}/{key...}", handlers.GetObject(s))
	mux.HandleFunc("DELETE /objects/{bucket}/{key...}", handlers.DeleteObject(s))

	log.Println("listening on :6060")
	log.Fatal(http.ListenAndServe(":6060", mux))
}
