package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	log.Println("Static files served at http://localhost:8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}
