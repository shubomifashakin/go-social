package routers

import (
	"database/sql"
	"net/http"

	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/handlers"
	"github.com/shubomifashakin/go-social/internal/mailer"
	"github.com/shubomifashakin/go-social/internal/middlewares"
	"go.uber.org/zap"
)

func CreateAuthRouter(m *http.ServeMux, db *sql.DB, cache *cache.Cache, mailer *mailer.Mailer, logger *zap.Logger, fromMail string) {
	authInstance:=&handlers.AuthHandler{	
		DB: db,
		Cache: cache,
		Mailer: mailer,
		Logger: logger,
		FromMail: fromMail,
	}
	
	isAuthorizedMware:=middlewares.IsAuthorized{Logger: logger}
	
	m.HandleFunc("POST /auth/sign-up",authInstance.SignUp)
	m.HandleFunc("POST /auth/login",authInstance.Login)
	m.HandleFunc("POST /auth/refresh",authInstance.Refresh)
	
	m.Handle("POST /auth/logout", isAuthorizedMware.Check(http.HandlerFunc(authInstance.Logout)))
	m.Handle("POST /auth/request-delete-account", isAuthorizedMware.Check(http.HandlerFunc(authInstance.RequestDelete)))
	m.Handle("DELETE /auth/me", isAuthorizedMware.Check(http.HandlerFunc(authInstance.DeleteMe)))
	m.Handle("GET /auth/me", isAuthorizedMware.Check(http.HandlerFunc(authInstance.GetMe)))
}