package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/render"
)

// Tipe context key untuk menghindari collision
type contextKey string

const (
	contextKeyUserID   contextKey = "userID"
	contextKeyUsername contextKey = "username"
)

// JWTAuthMiddleware memvalidasi token JWT dan menambahkan info user ke context
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// Coba ambil token dari cookie terlebih dahulu (metode yang diutamakan)
		cookieToken, err := GetSecureCookie(r, authTokenCookie)
		if err == nil && cookieToken != "" {
			tokenString = cookieToken
		} else {
			// Fallback ke Authorization header (untuk kompatibilitas ke belakang)
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}
		}

		// Jika token tidak ditemukan, return unauthorized
		if tokenString == "" {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication required. Please login.",
			})
			return
		}

		// Validasi token
		claims, err := ValidateJWT(tokenString)
		if err != nil {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
			})
			return
		}

		// Tambahkan info user ke context
		ctx := context.WithValue(r.Context(), contextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, contextKeyUsername, claims.Username)

		// Panggil handler berikutnya
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SecurityHeadersMiddleware menambahkan security headers
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Header keamanan
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Cek apakah ini route Swagger
		if strings.HasPrefix(r.URL.Path, "/swagger") {
			// Header yang lebih permisif untuk Swagger UI
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			// Izinkan inline scripts dan styles untuk Swagger UI
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:;")
		} else {
			// Header ketat untuk route API
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
		}

		// Panggil handler berikutnya
		next.ServeHTTP(w, r)
	})
}

