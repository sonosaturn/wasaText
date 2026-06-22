package main

import (
	"github.com/gorilla/handlers"
	"net/http"
)

func applyCORSHandler(h http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{
			"Content-Type",
			"Authorization", // Essential for the Bearer Token!
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT"}),
		// In development we use "*" or "http://localhost:5173"
		handlers.AllowedOrigins([]string{"*"}),
		handlers.MaxAge(86400), // CORS cache for 24 hours (avoids too many OPTIONS requests)
	)(h)
}
