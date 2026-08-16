package routers

import (
	"database/sql"
	"net/http"

	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/handlers"
	"github.com/shubomifashakin/go-social/internal/middlewares"
	"go.uber.org/zap"
)

func CreatePostRouter(m *http.ServeMux, db *sql.DB, cache *cache.Cache, logger *zap.Logger) {
	postsInstance:=&handlers.PostsHandler{	
		DB: db,
		Cache: cache,
		Logger: logger,
	}

	isAuthorizedMware:=middlewares.IsAuthorized{Logger: logger}

	m.Handle("POST /posts", isAuthorizedMware.Check(http.HandlerFunc(postsInstance.CreatePost)))
	m.Handle("DELETE /posts/{id}", isAuthorizedMware.Check(http.HandlerFunc(postsInstance.DeletePost)))
	m.Handle("GET /posts/{id}", isAuthorizedMware.Check(http.HandlerFunc(postsInstance.GetPost)))
}