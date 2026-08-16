package routers

import (
	"database/sql"
	"net/http"

	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/mailer"
	"go.uber.org/zap"
)

func RegisterRouter(db *sql.DB, redisInstance *cache.Cache, resend *mailer.Mailer, logger *zap.Logger,fromMail string) *http.ServeMux {
	mux := http.NewServeMux()
	
	v1 := http.NewServeMux()

	CreateAuthRouter(v1,db,redisInstance,resend,logger,fromMail)
	CreatePostRouter(v1,db,redisInstance,logger)

    mux.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))
	
	return mux
}