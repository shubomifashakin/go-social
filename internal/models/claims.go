package models

import "github.com/golang-jwt/jwt/v5"

type AccessTokenClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	jwt.RegisteredClaims
}