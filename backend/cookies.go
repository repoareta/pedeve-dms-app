package main

import (
	"net/http"
	"os"
)

const (
	authTokenCookie = "auth_token"
	cookieMaxAge    = 24 * 60 * 60 // 24 jam dalam detik
)

// SetSecureCookie mengatur cookie aman dengan flag yang sesuai
func SetSecureCookie(w http.ResponseWriter, name, value string) {
	isHTTPS := os.Getenv("ENV") == "production" || 
	           os.Getenv("HTTPS") == "true" ||
	           os.Getenv("FORCE_HTTPS") == "true"
	
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,              // Cegah serangan XSS
		Secure:   isHTTPS,           // Hanya kirim melalui HTTPS di production
		SameSite: http.SameSiteStrictMode, // Perlindungan CSRF
	})
}

// GetSecureCookie mengambil nilai cookie aman
func GetSecureCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// DeleteSecureCookie menghapus cookie aman
func DeleteSecureCookie(w http.ResponseWriter, name string) {
	isHTTPS := os.Getenv("ENV") == "production" || 
	           os.Getenv("HTTPS") == "true" ||
	           os.Getenv("FORCE_HTTPS") == "true"
	
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Hapus cookie
		HttpOnly: true,
		Secure:   isHTTPS,
		SameSite: http.SameSiteStrictMode,
	})
}

