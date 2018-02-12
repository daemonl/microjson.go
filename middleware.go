package microjson

import "net/http"

func VersionMiddleware(version string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Add("X-Version", version)
			next.ServeHTTP(rw, req)
		})
	}
}
