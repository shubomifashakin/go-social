package routers

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/handlers"
	"github.com/shubomifashakin/go-social/internal/mailer"
	"github.com/shubomifashakin/go-social/internal/middlewares"
	"github.com/shubomifashakin/go-social/internal/models"
	"go.uber.org/zap"
)

func CreateAuthRouter(m *http.ServeMux, usersRepo handlers.UsersRepo, cache *cache.Cache, mailer *mailer.Mailer, logger *zap.Logger, fromMail string) {
	authInstance:=&handlers.AuthHandler{	
		Cache: cache,
		Mailer: mailer,
		Logger: logger,
		FromMail: fromMail,
		UsersRepo: usersRepo,
	}
	
	isAuthorizedMware:=middlewares.IsAuthorized{Logger: logger}
	rl:=middlewares.RateLimiter{Store: cache,Logger: logger}
	
	m.HandleFunc("POST /auth/sign-up",authInstance.SignUp)
	m.Handle("POST /auth/login",rl.Check(http.HandlerFunc(authInstance.Login),func(r *http.Request) string {
		key:= r.Header.Get("X-Forwarded-For")
		if key != "" {
			firstIp:= strings.TrimSpace(strings.Split(key,",")[0])
			return "rate:login:"+firstIp
		}

		host,_,_:= net.SplitHostPort(r.RemoteAddr)
		return "rate:login:"+host
	},10,time.Minute))

	m.HandleFunc("POST /auth/refresh",authInstance.Refresh)
	
	m.Handle("POST /auth/logout", isAuthorizedMware.Check(http.HandlerFunc(authInstance.Logout)))
	m.Handle("POST /auth/request-delete-account", isAuthorizedMware.Check(
		rl.Check(http.HandlerFunc(authInstance.RequestDelete),func(r *http.Request) string {
			userInfo := r.Context().Value(middlewares.UserCtxKey).(models.UserRequestCtx)
			return "rate:delete-account-request:" + userInfo.UserId
		},10,time.Minute)),
	)
	m.Handle("DELETE /auth/me", isAuthorizedMware.Check(http.HandlerFunc(authInstance.DeleteMe)))
	m.Handle("GET /auth/me", isAuthorizedMware.Check(http.HandlerFunc(authInstance.GetMe)))
}