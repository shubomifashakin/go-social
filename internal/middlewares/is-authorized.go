package middlewares

import (
	"context"
	"net/http"

	"github.com/shubomifashakin/go-social/internal/models"
	"github.com/shubomifashakin/go-social/pkg/utils"
	"go.uber.org/zap"
)

type contextKey string
var UserCtxKey contextKey = "user"

type IsAuthorized struct {
	Logger *zap.Logger
}

func (i *IsAuthorized)Check(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get access token
		cookieAccessToken,err:= r.Cookie("access-token")
		if err != nil {
			i.Logger.Debug("Failed to get access token",zap.Error(err))

			utils.WriteResponse(w,http.StatusUnauthorized,models.MessageResponse{Message: "Unauthorized"})
			return
		}

		// verify the access token
		claims,err:=utils.VerifyAccessToken(cookieAccessToken.Value)
		if err != nil {
			i.Logger.Debug("Failed to verify access token",zap.Error(err))

			utils.WriteResponse(w,http.StatusUnauthorized,models.MessageResponse{Message: "Unauthorized"})
			return
		}

		// if the access token is valid, get the info we need and attach it to the request
		userId:=claims.Subject
		role:=claims.Role

		userInfo:= models.UserRequestCtx{
			UserId: userId,
			Role: role,
		}

		ctx:= context.WithValue(r.Context(),UserCtxKey,userInfo)

		n.ServeHTTP(w,r.WithContext(ctx))
	})
}