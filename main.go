package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

func main() {

	// Load settings:
	// ---
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}
	addr := ":" + port

	baseURL := os.Getenv("BASE_URL")
	if len(baseURL) == 0 {
		baseURL = fmt.Sprintf("http://localhost:%s", port)
	}

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
		BaseURL:    baseURL,
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

type Note struct {
	Data     []byte
	Destruct bool
}

type Server struct {
	BaseURL    string
	RedisCache *cache.Cache
}

func (s *Server) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
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

func (s *Server) notFound(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
}

func (s *Server) badRequest(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	message string,
) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}

func (s *Server) serverError(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Ops something went wrong. Please check the server logs."))
}

func (s *Server) renderTemplate(
	w http.ResponseWriter,
	r *http.Request,
	data interface{},
	name string,
	files ...string,
) {
	t := template.Must(template.ParseFiles(files...))
	err := t.ExecuteTemplate(w, name, data)
	if err != nil {
		panic(err)
	}
}

func (s *Server) renderMessage(
	w http.ResponseWriter,
	r *http.Request,
	title string,
	paragraphs ...interface{},
) {
	s.renderTemplate(
		w, r,
		struct {
			Title      string
			Paragraphs []interface{}
		}{
			Title:      title,
			Paragraphs: paragraphs,
		},
		"layout",
		"dist/layout.html",
		"dist/message.html",
	)
}

func (s *Server) handlePOST(
	w http.ResponseWriter,
	r *http.Request,
) {
	mediaType := r.Header.Get("Content-Type")
	if mediaType != "application/x-www-form-urlencoded" {
		s.badRequest(
			w, r,
			http.StatusUnsupportedMediaType,
			"Invalid media type posted.")
		return
	}

	err := r.ParseForm()
	if err != nil {
		s.badRequest(
			w, r,
			http.StatusBadRequest,
			"Invalid form data posted.")
		return
	}
	form := r.PostForm

	message := form.Get("message")
	destruct := false
	ttl := time.Hour * 24
	if form.Get("ttl") == "untilRead" {
		destruct = true
		ttl = ttl * 365
	}

	note := &Note{
		Data:     []byte(message),
		Destruct: destruct,
	}

	key := uuid.NewString()
	err = s.RedisCache.Set(
		&cache.Item{
			Ctx:            r.Context(),
			Key:            key,
			Value:          note,
			TTL:            ttl,
			SkipLocalCache: true,
		})
	if err != nil {
		fmt.Println(err)
		s.serverError(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	noteURL := fmt.Sprintf("%s/%s", s.BaseURL, key)
	s.renderMessage(
		w, r,
		"Note was successfully created",
		template.HTML(
			fmt.Sprintf("<a href='%s'>%s</a>", noteURL, noteURL)))
}

func (s *Server) handleGET(
	w http.ResponseWriter,
	r *http.Request,
) {
	path := r.URL.Path
	if path == "/" {
		s.renderTemplate(w, r, nil, "layout", "dist/layout.html", "dist/index.html")
		return
	}

	noteID := strings.TrimPrefix(path, "/")

	ctx := r.Context()
	note := &Note{}
	err := s.RedisCache.GetSkippingLocalCache(
		ctx,
		noteID,
		note)
	if err != nil {
		s.badRequest(
			w, r,
			http.StatusNotFound,
			fmt.Sprintf("Note with ID %s does not exist.", noteID))
		return
	}

	if note.Destruct {
		err := s.RedisCache.Delete(ctx, noteID)
		if err != nil {
			fmt.Println(err)
			s.serverError(w, r)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write(note.Data)
}
