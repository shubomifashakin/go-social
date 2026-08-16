package routers

import (
	"database/sql"
	"net/http"

	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/handlers"
	"github.com/shubomifashakin/go-social/internal/mailer"
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

	m.HandleFunc("POST /auth/sign-up",authInstance.SignUp)
	m.HandleFunc("POST /auth/login",authInstance.Login)
}