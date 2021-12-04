package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}
	addr := ":" + port

	fmt.Printf("Starting web server, listening on %s\n", addr)

	err := http.ListenAndServe(addr, &Server{})
	if err != nil {
		panic(err)
	}
}

type Server struct{}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" || r.Method == "HEAD" {
		noteID := strings.TrimPrefix(r.URL.Path, "/")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("You requested the note with the ID '%s'.", noteID)))
		return
	}

	if r.Method == "POST" && r.URL.Path == "/notes" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("You posted to /notes."))
		return
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
}
