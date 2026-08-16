package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/models"
	"github.com/shubomifashakin/go-social/pkg/utils"
	"go.uber.org/zap"
)

type RateLimiter struct {
	Store *cache.Cache
	Logger *zap.Logger
}

func (rl *RateLimiter) Check(next http.Handler, keyFunc func(*http.Request) string, limit int64, window time.Duration) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()

        key := keyFunc(r)

        limited,expiresAt, err := rl.Store.IsRateLimited(ctx, key, limit, window)
        if err != nil {
            rl.Logger.Error("Failed to check rate limit", zap.Error(err))
            utils.WriteResponse(w, http.StatusInternalServerError, models.MessageResponse{Message: "Internal server error"})
            return
        }

        if limited {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Until(expiresAt).Seconds())))
			utils.WriteResponse(w, http.StatusTooManyRequests, models.MessageResponse{Message: "Too many requests"})
			return
		}

        next.ServeHTTP(w, r)
    })
}
