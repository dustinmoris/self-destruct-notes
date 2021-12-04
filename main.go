package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}
	addr := ":" + port

	fmt.Printf("Starting web server, listening on %s\n", addr)

	err := http.ListenAndServe(addr, http.HandlerFunc(webServer))
	if err != nil {
		panic(err)
	}
}

func webServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("hello world"))
}
