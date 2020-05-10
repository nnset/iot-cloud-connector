package servers

import (
	"net/http"
)

/*
APIAuthMiddleWare TODO
*/
type APIAuthMiddleWare interface {
	Middleware(next http.Handler) http.Handler
}

/*
APIStaticBasicAuthenticationMiddleware TODO
*/
type APIStaticBasicAuthenticationMiddleware struct {
	Username string
	Password string
}

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

/*
APINoAuthenticationMiddleware TODO
*/
type APINoAuthenticationMiddleware struct {
}

func (na *APINoAuthenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
