package middleware

import (
	"net/http"

	"github.com/lfkeitel/golang-app-framework/src/utils"
)

func SetSessionInfo(e *utils.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := e.Sessions.GetSession(r)
		r = utils.SetSessionToContext(r, session)
		r = utils.SetEnvironmentToContext(r, e)

		// If running behind a proxy, set the RemoteAddr to the real address
		if r.Header.Get("X-Real-IP") != "" {
			r.RemoteAddr = r.Header.Get("X-Real-IP")
		}
		r = utils.SetIPToContext(r)

		next.ServeHTTP(w, r)
	})
}
