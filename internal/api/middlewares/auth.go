package middlewares

import (
	"context"
	"net/http"
)

func (m Middlewares) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("unauthorized"))
			return
		}
		user, err := m.authService.ValidateToken(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("unauthorized"))
			return
		}
		m.logger.Log.Debug("middleware authenticated username: ", user.Username)

		ctx := context.WithValue(r.Context(), "username", user.Username)
		ctx = context.WithValue(ctx, "uuid", user.UUID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
