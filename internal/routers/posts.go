package routers

import (
	"net/http"
	"time"

	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/handlers"
	"github.com/shubomifashakin/go-social/internal/middlewares"
	"github.com/shubomifashakin/go-social/internal/models"
	"go.uber.org/zap"
)

func CreatePostRouter(m *http.ServeMux, postsRepo handlers.PostsRepo, cache cache.CacheService, logger *zap.Logger) {
	postsInstance:=&handlers.PostsHandler{	
		Cache: cache,
		Logger: logger,
		PostsRepo: postsRepo,
	}

	isAuthorizedMware:=middlewares.IsAuthorized{Logger: logger}
	rl:= middlewares.RateLimiter{Store: cache,Logger: logger}

	m.Handle("POST /posts", isAuthorizedMware.Check(
		rl.Check(
			http.HandlerFunc(postsInstance.CreatePost),
			func(r *http.Request) string {
				userInfo := r.Context().Value(middlewares.UserCtxKey).(models.UserRequestCtx)
				return "rate:create-post:" + userInfo.UserId
			},
			30, 1*time.Minute,
		),
	))

	m.Handle("GET /posts", isAuthorizedMware.Check(http.HandlerFunc(postsInstance.GetPosts)))
	m.Handle("DELETE /posts/{id}", isAuthorizedMware.Check(
		rl.Check(http.HandlerFunc(postsInstance.DeletePost), func(r *http.Request) string {
			userInfo := r.Context().Value(middlewares.UserCtxKey).(models.UserRequestCtx)
			return "rate:delete-post:" + userInfo.UserId
		}, 30, 1*time.Minute),
	))

	m.Handle("GET /posts/{id}", isAuthorizedMware.Check(http.HandlerFunc(postsInstance.GetPost)))
	m.Handle("GET /users/{id}/posts", isAuthorizedMware.Check(http.HandlerFunc(postsInstance.GetPostsForUser)))
}