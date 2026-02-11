package main

import (
	"github.com/gorilla/handlers"
	"net/http"
)

func applyCORSHandler(h http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{
			"Content-Type",
			"Authorization", // Fondamentale per il Bearer Token!
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT"}),
		// In sviluppo usiamo "*" o "http://localhost:5173"
		handlers.AllowedOrigins([]string{"*"}), 
		handlers.MaxAge(86400), // Cache del CORS per 24 ore (evita troppe richieste OPTIONS)
	)(h)
}