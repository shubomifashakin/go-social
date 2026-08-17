package handlers

import (
	"context"
	"time"

	"github.com/shubomifashakin/go-social/internal/models"
)

type PostsRepo interface {
	CreatePost(ctx context.Context, userId string, post models.CreatePost) (string, error)
	DeletePostById(ctx context.Context, postId string, userId string) error
	GetPostById(ctx context.Context, postId string) (models.Post, error)
	GetPaginatedPostsForUser(ctx context.Context, userId string, limit int, cursor string) ([]models.Post, error)
}

type UsersRepo interface {
	CreateUser(ctx context.Context, user models.UserSignup) (string, error)
	DeleteUserAccountById(ctx context.Context, id string) error
	FindUserByUsername(ctx context.Context, username string) (models.User, error)
	FindUserById(ctx context.Context, userId string) (models.User, error)
	CreateRefreshToken(ctx context.Context, userId string, tokenId string, expiresAt time.Time) error
	FindRefreshTokenByTokenId(ctx context.Context, tokenId string) (models.RefreshToken, error)
	RotateRefreshToken(ctx context.Context, userId string, oldTokenId string, newTokenId string, expiresAt time.Time) error
	DeleteRefreshTokenByTokenId(ctx context.Context, tokenId string) error
}

