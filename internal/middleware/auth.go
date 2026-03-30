package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"noc-mcp/pkg/logger"

	"go.uber.org/zap"
)

func BearerAuth(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			logger.Log.Warn("request sin header Authorization",
				zap.String("remote", r.RemoteAddr),
				zap.String("path", r.URL.Path),
			)
			http.Error(w, `{"error":"unauthorized","message":"Bearer token requerido"}`, http.StatusUnauthorized)
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			http.Error(w, `{"error":"unauthorized","message":"esquema de auth inválido, use Bearer"}`, http.StatusUnauthorized)
			return
		}

		token := auth[len(prefix):]
		if subtle.ConstantTimeCompare([]byte(token), []byte(apiKey)) != 1 {
			logger.Log.Warn("token inválido rechazado",
				zap.String("remote", r.RemoteAddr),
				zap.String("path", r.URL.Path),
			)
			http.Error(w, `{"error":"forbidden","message":"token inválido"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
