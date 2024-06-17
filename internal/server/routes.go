package server

import (
	"encoding/json"
	"github.com/a-h/templ"
	"github.com/mdhender/otto/frontend"
	"log"
	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {
	// default mux, no routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.HelloWorldHandler)

	mux.HandleFunc("/health", s.healthHandler)

	fileServer := http.FileServer(http.FS(frontend.Files))
	mux.Handle("/assets/", fileServer)
	mux.Handle("/features", templ.Handler(frontend.FeaturesPage()))
	mux.Handle("/landing", templ.Handler(frontend.LandingPage()))
	mux.Handle("/web", templ.Handler(frontend.HelloForm()))
	mux.HandleFunc("/hello", frontend.HelloWebHandler)

	return mux
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}
