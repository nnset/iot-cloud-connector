package servers

import (
	"net/http"
)

// CrossOriginResourceSharing CORS settings for API server
type CrossOriginResourceSharing struct {
	Origin  string
	Headers string
}

// APIAuthMiddleWare Use this interface to code how REST API requests
// should be authenticated.
// This interface is compatible with Gorilla Mux Router (http://www.gorillatoolkit.org/pkg/mux).
type APIAuthMiddleWare interface {
	Middleware(next http.Handler) http.Handler
}

// APIStaticBasicAuthenticationMiddleware Authenticate REST API requests with a unique
// username and password combination.
type APIStaticBasicAuthenticationMiddleware struct {
	Username string
	Password string
}

// Middleware Gorilla Mux Router middleware method for APIStaticBasicAuthenticationMiddleware
func (ba *APIStaticBasicAuthenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		username, password, authOK := r.BasicAuth()

		if authOK == false || username != ba.Username || password != ba.Password {
			http.Error(w, "Not authorized", 401)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// APINoAuthenticationMiddleware No authentication required for API REST requests.
type APINoAuthenticationMiddleware struct {
}

// Middleware Gorilla Mux Router middleware method for APINoAuthenticationMiddleware
func (na *APINoAuthenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
