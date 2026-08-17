package routers

import (
	"database/sql"
	"net/http"

	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/mailer"
	"github.com/shubomifashakin/go-social/internal/middlewares"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	_ "github.com/shubomifashakin/go-social/docs"
)

func RegisterRouter(environment string, db *sql.DB, redisInstance *cache.Cache, resend *mailer.Mailer, logger *zap.Logger,fromMail string) *http.ServeMux {
	mux := http.NewServeMux()
	
	v1 := http.NewServeMux()

	rLogger:= middlewares.RequestLogger{Logger: logger}

	CreateAuthRouter(v1,db,redisInstance,resend,logger,fromMail)
	CreatePostRouter(v1,db,redisInstance,logger)

    mux.Handle("/api/v1/", http.StripPrefix("/api/v1", rLogger.Check(v1)))
	
	if environment !="production" {
		mux.Handle("/swagger/", httpSwagger.WrapHandler)
	}
	
	return mux
}