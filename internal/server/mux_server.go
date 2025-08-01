package server

import (
	"fmt"
	"net/http"
	"twitter-clone/internal/domain/config"

	"github.com/gorilla/mux"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func NewMuxServer(config config.APIConfig) (*http.Server, *mux.Router) {
	router := mux.NewRouter()

	// Wrap router with CORS middleware
	handler := enableCORS(router)

	commonAddress := fmt.Sprintf("%s:%d", config.Host(), config.Port())
	return &http.Server{
		Addr:    commonAddress, // Configurable port
		Handler: handler,
	}, router
}
