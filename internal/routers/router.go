package routers

import (
	"net/http"

	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/handlers"
	"github.com/shubomifashakin/go-social/internal/mailer"
	"github.com/shubomifashakin/go-social/internal/middlewares"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	_ "github.com/shubomifashakin/go-social/docs"
)

func RegisterRouter(environment string, usersRepo handlers.UsersRepo, postsRepo handlers.PostsRepo, redisInstance *cache.Cache, resend *mailer.Mailer, logger *zap.Logger, fromMail string) *http.ServeMux {
	mux := http.NewServeMux()
	
	v1 := http.NewServeMux()

	rLogger:= middlewares.RequestLogger{Logger: logger}
	
	CreateAuthRouter(v1,usersRepo,redisInstance,resend,logger,fromMail)
	CreatePostRouter(v1,postsRepo,redisInstance,logger)

    mux.Handle("/api/v1/", http.StripPrefix("/api/v1", rLogger.Check(v1)))

	if environment !="production" {
		mux.Handle("/swagger/", httpSwagger.WrapHandler)
	}
	
	return mux
}