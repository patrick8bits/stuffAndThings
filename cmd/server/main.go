package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./web"))

	mux.Handle("/", fileServer)

	mux.HandleFunc("/api/sort", sortHandler)
	mux.HandleFunc("/api/sort/source", sortSourceHandler)

	log.Println("Server starting at http://localhost:8080")

	err := http.ListenAndServe("127.0.0.1:8080", mux)

	if err != nil {
		log.Fatal(err)
	}
}
