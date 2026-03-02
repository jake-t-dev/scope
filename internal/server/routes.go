package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", s.HandleGitHubQuery)

	return r
}

func (s *Server) HandleGitHubQuery(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	interests, err := s.ghClient.GetUserInterests(r.Context(), username)
	if err != nil {
		log.Printf("error fetching GitHub data. Err: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	news, err := s.newsClient.GetTailoredNews(r.Context(), interests)
	if err != nil {
		log.Printf("error fetching news data. Err: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	jsonResp, err := json.Marshal(news)
	if err != nil {
		log.Printf("error handling JSON marshal. Err: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(jsonResp)
}
