package main

import (
	"log"
	"net/http"
)

func main() {
	fileServer := http.FileServer(http.Dir("./web"))

	http.Handle("/", fileServer)

	log.Println("Server starting at http://localhost:8080")

	err := http.ListenAndServe("127.0.0.1:8080", nil)

	if err != nil {
		log.Fatal(err)
	}
}
