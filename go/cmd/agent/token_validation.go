package main

import (
	"net/http"
	"strings"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Entering authMiddleware")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		token := headerParts[1]

		// TODO: validate the token

		logger.Sugar().Debugf("Your token: %s", token)

		//TODO
		// Add auth value to request context for later use in handlers if needed
		// ctx := context.WithValue(r.Context(), "auth", auth)
		// r = r.WithContext(ctx)
		logger.Info("Serve next")
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
