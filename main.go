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
		s.handleGET(w, r)
		return
	}
	if r.Method == "POST" && r.URL.Path == "/" {
		s.handlePOST(w, r)
		return
	}
	s.notFound(w, r)
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
}

func (s *Server) handlePOST(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("You posted to /."))
}

func (s *Server) handleGET(w http.ResponseWriter, r *http.Request) {
	noteID := strings.TrimPrefix(r.URL.Path, "/")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("You requested the note with the ID '%s'.", noteID)))
}
