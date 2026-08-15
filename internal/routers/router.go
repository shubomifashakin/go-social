package routers

import (
	"database/sql"
	"net/http"

	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/mailer"
	"go.uber.org/zap"
)

func RegisterRouter(db *sql.DB, redisInstance *cache.Cache, resend *mailer.Mailer, logger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	return mux
}