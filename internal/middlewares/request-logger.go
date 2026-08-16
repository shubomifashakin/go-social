package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RequestLogger struct {
	Logger *zap.Logger
}

type responseWriter struct {
    http.ResponseWriter
    status int
}

var RequestIDKey contextKey = "request-id"

func (rw *responseWriter) WriteHeader(code int) {
    rw.status = code
    rw.ResponseWriter.WriteHeader(code)
}

func (rl *RequestLogger)Check(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now:=time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		requestId:=uuid.NewString()
		ctx:=context.WithValue(r.Context(),RequestIDKey,requestId)
		w.Header().Set("X-Request-Id", requestId)

		next.ServeHTTP(rw,r.WithContext(ctx))

		timeTaken:=time.Since(now)

		rl.Logger.Info("Request",
			zap.String("method",r.Method),
			zap.String("path",r.Pattern),
			zap.Int("status",rw.status),
			zap.Duration("latency",timeTaken),
			zap.String("request-id",requestId),
		)
	})
}