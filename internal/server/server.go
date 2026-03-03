package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"scope/internal/ai"
	"scope/internal/github"
	"scope/internal/news"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port       int
	ghClient   github.Client
	newsClient news.Client
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	aiClient, err := ai.NewClient(os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		fmt.Printf("Warning: AI client initialization failed: %v. Search results will be limited.\n", err)
	}

	NewServer := &Server{
		port:       port,
		ghClient:   github.NewClient(),
		newsClient: news.NewClient(aiClient),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
