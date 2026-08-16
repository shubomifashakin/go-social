package utils

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shubomifashakin/go-social/internal/models"
)

func SignToken(claims jwt.Claims) (string,error) {
	key:= os.Getenv("JWT_PRIVATE_KEY")
	if key == "" {
		return "",fmt.Errorf("JWT Secret Not Configured")
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedAccessToken, err := accessToken.SignedString([]byte(key))

	return signedAccessToken,err
}


func VerifyAccessToken(tokenString string) (*jwt.Token,error){
	key:= os.Getenv("JWT_PRIVATE_KEY")
	
	if key == "" {
		return nil,fmt.Errorf("JWT Secret Not Configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &models.AccessTokenClaims{},func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid signing method: %v", token.Header["alg"])
		}
		return []byte(key), nil
	})
	

	return token,err
}

func VerifyRefreshToken(tokenString string) (*jwt.Token,error){
	key:= os.Getenv("JWT_PRIVATE_KEY")
	
	if key == "" {
		return nil,fmt.Errorf("JWT Secret Not Configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &models.RefreshTokenClaims{},func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid signing method: %v", token.Header["alg"])
		}
		return []byte(key), nil
	})
	

	return token,err
}