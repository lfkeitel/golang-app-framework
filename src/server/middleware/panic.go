package middleware

import (
	"net/http"
	"runtime"

	"github.com/lfkeitel/golang-app-framework/src/utils"
	"github.com/lfkeitel/verbose"
)

func Panic(e *utils.Environment, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(runtime.Error); ok {
					buf := make([]byte, 2048)
					runtime.Stack(buf, false)
					e.Log.WithFields(verbose.Fields{
						"package": "middleware:panic",
						"stack":   string(buf),
					}).Alert()
				}
				e.Log.Alert(r)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
