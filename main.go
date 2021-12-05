package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

func main() {

	// Load settings:
	// ---
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}
	addr := ":" + port

	redisURL := os.Getenv("REDIS_URL")
	if len(redisURL) == 0 {
		redisURL = "redis://:@localhost:6379/1"
	}

	// Bootstrap:
	// ---
	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}
	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()
	redisCache := cache.New(&cache.Options{
		Redis: redisClient,
	})
	server := &Server{
		RedisCache: redisCache,
	}

	// Start web server:
	// ---
	fmt.Printf("Starting web server, listening on %s\n", addr)
	err = http.ListenAndServe(addr, server)
	if err != nil {
		panic(err)
	}
}

type Server struct {
	RedisCache *cache.Cache
}

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
