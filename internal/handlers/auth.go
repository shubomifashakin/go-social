package handlers

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/mailer"
	"github.com/shubomifashakin/go-social/internal/middlewares"
	"github.com/shubomifashakin/go-social/internal/models"
	"github.com/shubomifashakin/go-social/internal/repository"
	"github.com/shubomifashakin/go-social/internal/templates"
	"github.com/shubomifashakin/go-social/pkg/utils"
	"go.uber.org/zap"
)


type AuthHandler struct {
	DB *sql.DB
	Cache *cache.Cache
	Mailer *mailer.Mailer
	Logger *zap.Logger
	FromMail string
}

func (a *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request){
	ctx,cancel:= context.WithTimeout(r.Context(),10*time.Second)
	defer cancel()

	// extract the post body
	var body models.UserSignup
	
	if err:=json.NewDecoder(r.Body).Decode(&body); err != nil {
		a.Logger.Debug("Failed to parse sign up body",zap.Error(err))
		
		utils.WriteResponse(w,http.StatusBadRequest,models.MessageResponse{
			Message: "Malformed Body",
		})

		return
	}

	if err := utils.Validator.Struct(body); err != nil {
		var validationErrors validator.ValidationErrors

		if errors.As(err, &validationErrors) {
			fields := make(map[string]string)

			for _, e := range validationErrors {
				fields[e.Field()] = e.Tag() 
			}
			
			utils.WriteResponse(w,http.StatusBadRequest,fields)
			return
		}
	}

	hash,err:= utils.HashPassword(body.Password)
	if err != nil {
		a.Logger.Error("Failed to hash password",zap.Error(err))

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal server error"})
		return
	}

	body.Password=hash

	_,err= repository.CreateUser(ctx,a.DB,body)
	
	if err != nil {
		switch {
			case errors.Is(err, models.ErrDuplicateEntry):
				utils.WriteResponse(w,http.StatusConflict,models.MessageResponse{
					Message:"Email or Username already taken",
				})

			case errors.Is(err, models.ErrMissingField):
				utils.WriteResponse(w,http.StatusBadRequest,models.MessageResponse{
					Message:"Invalid payload",
				})

			default:
				utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{
						Message:"Internal server error",
					})
		}

		return
	}
	
	var builder strings.Builder

	tmpl, err := template.New("welcome-mail").Parse(templates.WelcomeTemplate)
	if err != nil {
		a.Logger.Error("Failed to parse welcome template", zap.Error(err))
	} else {
		err = tmpl.Execute(&builder, body)
		if err != nil {
			a.Logger.Error("Failed to build welcome email template", zap.Error(err))
		}
	}

	if err == nil {
		_,err=a.Mailer.SendMail(mailer.Mail{
			From:a.FromMail ,
			To: []string{body.Email},
			Subject: "Welcome to GO-Social",
			Html: builder.String(),
		})

		if err != nil {
			a.Logger.Error("Failed to send mail",zap.Error(err))
		}
	}
	
	utils.WriteResponse(w,http.StatusCreated,models.MessageResponse{
		Message:"Success",
	})
}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request){
	ctx,cancel:= context.WithTimeout(r.Context(),10*time.Second)
	defer cancel()

	// extract the post body
	var body models.UserLogin
	
	if err:=json.NewDecoder(r.Body).Decode(&body); err != nil {
		a.Logger.Debug("Failed to parse sign up body",zap.Error(err))
		
		utils.WriteResponse(w,http.StatusBadRequest,models.MessageResponse{
			Message: "Invalid Body",
		})

		return
	}

	// validate the payload sent
	if err := utils.Validator.Struct(body); err != nil {
		var validationErrors validator.ValidationErrors

		if errors.As(err, &validationErrors) {
			fields := make(map[string]string)

			for _, e := range validationErrors {
				fields[e.Field()] = e.Tag() 
			}
			
			utils.WriteResponse(w,http.StatusBadRequest,fields)
			return
		}
	}

	// find the user with the username and if they dont exist, return not found
	user,err:= repository.FindUserByUsername(ctx,a.DB,body.Username)
	if err != nil {
		switch err {
			case models.ErrNotFound:
				a.Logger.Debug(fmt.Sprintf("%s does not exist",body.Username))

				utils.WriteResponse(w,http.StatusNotFound,models.MessageResponse{Message: "User does not exist"})
				return
			default:
				a.Logger.Error("Failed to get user",zap.Error(err))

				utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{
					Message:"Internal server error",
				})
				return

		}
	}

	// compare the password sent with the one stored in the database
	if isTheSame:= utils.VerifyPassword(body.Password,user.Password); !isTheSame {
		a.Logger.Debug(fmt.Sprintf("%s is not authorized",body.Username))

		utils.WriteResponse(w,http.StatusUnauthorized,models.MessageResponse{Message: "Unauthorized"})

		return
	}

	// generate the jwts
	accessTokenId:= uuid.NewString()
	refreshTokenId:= uuid.NewString()

	accessTokenExpiresAt:= time.Now().Add(10*time.Minute)
	refreshExpiresAt:= time.Now().Add(14*24*time.Hour)

	accessClaims:=models.AccessTokenClaims{
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: user.ID,
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiresAt),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ID: accessTokenId,
		},
	}

	refreshClaims:= models.RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{	Subject: user.ID,
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ID: refreshTokenId,
			},
	}

	accessToken,accessErr:=utils.SignToken(accessClaims)
	refreshToken,refreshErr:=utils.SignToken(refreshClaims)

	if accessErr != nil || refreshErr!= nil {
		if accessErr != nil {
			a.Logger.Error("Failed to generate access token",zap.Error(accessErr))
		}else{
			a.Logger.Error("Failed to generate refresh token",zap.Error(refreshErr))
		}

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal server error"})

		return
	}

	// create the refresh token in the db
	err=repository.CreateRefreshToken(ctx,a.DB,user.ID,refreshTokenId,refreshExpiresAt)

	if err !=nil {
		a.Logger.Error("Failed to store refresh token in db",zap.Error(err))

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal server error"})
		return
	}

	// set the access and refresh token as cookies
	http.SetCookie(w,&http.Cookie{
		Name: "access-token",
		HttpOnly: true,
		Domain: "localhost",
		Path: "/",
		Secure: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge: int(time.Until(accessTokenExpiresAt).Seconds()),
		Expires: accessTokenExpiresAt,
		Value: accessToken,
	})
	
	http.SetCookie(w,&http.Cookie{
		Name: "refresh-token",
		HttpOnly: true,
		Domain: "localhost",
		Path: "/",
		Secure: true,
		SameSite: http.SameSiteLaxMode,
		Expires:refreshExpiresAt,
		MaxAge: int(time.Until(refreshExpiresAt).Seconds()),
		Value: refreshToken,
	})

	utils.WriteResponse(w,http.StatusOK,models.MessageResponse{Message: "Success"})
}

func (a *AuthHandler) Logout(w http.ResponseWriter, r *http.Request){
	ctx,cancel:= context.WithTimeout(r.Context(),10*time.Second)
	defer cancel()

	// get the refresh token from the cookie
	cookieRefreshToken,err:= r.Cookie("refresh-token")
	if err != nil {
		a.Logger.Debug("Failed to get refresh token from the cookies",zap.Error(err))

		utils.WriteResponse(w,http.StatusUnauthorized,models.MessageResponse{Message: "Unauthorized"})
		return
	}
	
	// extract the refreshtokenId from the token
	claims,err:= utils.VerifyRefreshToken(cookieRefreshToken.Value)
	if err != nil {
		a.Logger.Debug("Refresh token verification failed",zap.Error(err))

		utils.WriteResponse(w,http.StatusUnauthorized,models.MessageResponse{Message: "Unauthorized"})
		return
	}
	
	// delete the refresh token with that id from the db
	err= repository.DeleteRefreshTokenByTokenId(ctx,a.DB,claims.ID)
	if err != nil {
		a.Logger.Error("Failed to delete refresh token",zap.Error(err))

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal Server Error"})
		return
	}

	// clear the cookies by setting MaxAge to -1 and expiring them in the past
	http.SetCookie(w, &http.Cookie{
		Name:     "access-token",
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
		Domain: "localhost",
		Path: "/",
		Expires:  time.Unix(0, 0),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh-token",
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
		Domain: "localhost",
		Path: "/",
		Expires:  time.Unix(0, 0),
	})

	utils.WriteResponse(w,http.StatusOK,models.MessageResponse{
		Message: "Success",
	})
}

func (a *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request){
	ctx,cancel:= context.WithTimeout(r.Context(),10*time.Second)
	defer cancel()

	// get the refresh token from the cookies
	cookieRefreshToken,err:= r.Cookie("refresh-token")

	if err != nil {
		a.Logger.Debug("Refresh token does not exist")

		utils.WriteResponse(w, http.StatusUnauthorized, models.MessageResponse{Message: "Unauthorized"})
		return
	}

	// verify the refresh token
	verifiedToken,err:= utils.VerifyRefreshToken(cookieRefreshToken.Value)

	if err != nil {
		a.Logger.Debug("Refresh token is invalid")

		utils.WriteResponse(w, http.StatusUnauthorized, models.MessageResponse{Message: "Unauthorized"})
		return
	}

	// get the refresh token id from the jwt
	tokenId:= verifiedToken.ID

	// get the refresh token from the db and check if it has expired
	oldRefreshToken,err:= repository.FindRefreshTokenByTokenId(ctx,a.DB,tokenId)

	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			utils.WriteResponse(w, http.StatusUnauthorized, models.MessageResponse{Message: "Unauthorized"})
		default:
			a.Logger.Error("Failed to find refresh token", zap.Error(err))

			utils.WriteResponse(w, http.StatusInternalServerError, models.MessageResponse{Message: "Internal server error"})
		}
		return
	}	

	if oldRefreshToken.ExpiresAt.Before(time.Now()) {
		a.Logger.Debug("Refresh token has expired")

		utils.WriteResponse(w, http.StatusUnauthorized,models.MessageResponse{Message: "Unauthorized"})
		return
	}

	// get the users info from the db
	user,err:= repository.FindUserById(ctx,a.DB,oldRefreshToken.UserID)

	if err != nil {
		a.Logger.Error("Failed to get user info",zap.Error(err))

		utils.WriteResponse(w, http.StatusInternalServerError,models.MessageResponse{Message: "Internal server error"})
		return	
	}

	// generate the jwts
	accessTokenId:= uuid.NewString()
	refreshTokenId:= uuid.NewString()

	accessTokenExpiresAt:= time.Now().Add(10*time.Minute)
	refreshExpiresAt:= time.Now().Add(14*24*time.Hour)

	accessClaims:=models.AccessTokenClaims{
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: user.ID,
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiresAt),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ID: accessTokenId,
		},
	}

	refreshClaims:= models.RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{	Subject: user.ID,
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ID: refreshTokenId,
			},
	}

	accessToken,accessErr:=utils.SignToken(accessClaims)
	refreshToken,refreshErr:=utils.SignToken(refreshClaims)

	if accessErr != nil || refreshErr!= nil {
		if accessErr != nil {
			a.Logger.Error("Failed to generate access token",zap.Error(accessErr))
		}else{
			a.Logger.Error("Failed to generate refresh token",zap.Error(refreshErr))
		}

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal server error"})

		return
	}

	// rotate the refresh token
	err=repository.RotateRefreshToken(ctx,a.DB,user.ID,oldRefreshToken.TokenID,refreshTokenId,refreshExpiresAt)

	if err !=nil {
		a.Logger.Error("Failed to rotate refresh token",zap.Error(err))

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal server error"})
		return
	}

	// set the access and refresh token as cookies
	http.SetCookie(w,&http.Cookie{
		Name: "access-token",
		HttpOnly: true,
		Domain: "localhost",
		Path: "/",
		Secure: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge: int(time.Until(accessTokenExpiresAt).Seconds()),
		Expires: accessTokenExpiresAt,
		Value: accessToken,
	})

	http.SetCookie(w,&http.Cookie{
		Name: "refresh-token",
		HttpOnly: true,
		Domain: "localhost",
		Path: "/",
		Secure: true,
		SameSite: http.SameSiteLaxMode,
		Expires:refreshExpiresAt,
		MaxAge: int(time.Until(refreshExpiresAt).Seconds()),
		Value: refreshToken,
	})

	// return the response to the user, setting the cookies
	utils.WriteResponse(w,http.StatusOK, models.MessageResponse{Message: "Success"})
}

func (a *AuthHandler) RequestDelete(w http.ResponseWriter, r * http.Request){
	ctx,cancel:= context.WithTimeout(r.Context(),10*time.Second)
	defer cancel()

	// get the user context from the request
	userCtxInfo,ok:= r.Context().Value(middlewares.UserCtxKey).(models.UserRequestCtx)
	if !ok {
		a.Logger.Debug("User is unauthorized")

		utils.WriteResponse(w,http.StatusUnauthorized,models.MessageResponse{Message: "Unauthorized"})
		return
	}

	// get the users info from the user
	user,err:=repository.FindUserById(ctx,a.DB,userCtxInfo.UserId)
	if err != nil {
		a.Logger.Error("Failed to get user info",zap.Error(err))

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal Server Error"})
		return
	}

	// generate an access code for the user
	code:= utils.GenerateSixDigitCode()

	deleteCodeInfo:= models.DeleteCode{
		Code: code,
	}

	// store the access code in the cache
	err=a.Cache.SetJSON(ctx,fmt.Sprintf("user:%s:delete-request-code",user.ID),deleteCodeInfo,5*time.Minute)

	if err != nil {
		a.Logger.Error("Failed to set delete code in cache",zap.Error(err))

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal Server Error"})
		return
	}

	var builder strings.Builder

	// generate the html 
	tmpl,err:=template.New("delete-request-mail").Parse(templates.DeleteRequestTemplate)
	if err != nil {
		a.Logger.Error("Failed to generate delete html",zap.Error(err))

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal Server Error"})
		return
	}

	deleteInfo:=struct{
		FirstName string
		Code string
	}{
		FirstName: user.FirstName,
		Code: code,
	}

	if err=tmpl.Execute(&builder,deleteInfo); err!=nil {
		a.Logger.Error("Failed to execute delete html templates",zap.Error(err))

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal Server Error"})
		return
	}
	
	// send the mail to the user
	_,err=a.Mailer.SendMail(mailer.Mail{
		From: a.FromMail,
		To: []string{user.Email},
		Subject: "Delete Code",
		Html: builder.String(),
	})

	if err != nil {
		a.Logger.Error("Failed to send delete mail",zap.Error(err))

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal Server Error"})
		return
	}

	// return a response to the user
	utils.WriteResponse(w,http.StatusOK,models.MessageResponse{Message: "Success"})	
}

func (a *AuthHandler) DeleteMe(w http.ResponseWriter, r *http.Request){
	ctx,cancel:= context.WithTimeout(r.Context(),10*time.Second)
	defer cancel()

	// get the user context from the request
	userCtxInfo,ok:= r.Context().Value(middlewares.UserCtxKey).(models.UserRequestCtx)
	if !ok {
		a.Logger.Debug("User is unauthorized")

		utils.WriteResponse(w,http.StatusUnauthorized,models.MessageResponse{Message: "Unauthorized"})
		return
	}

	// get the post body from the request body
	var body models.DeleteCode

	err:=json.NewDecoder(r.Body).Decode(&body)
	if err !=nil {
		a.Logger.Debug("Invalid post body",zap.Error(err))

		utils.WriteResponse(w,http.StatusBadRequest,models.MessageResponse{Message: "Invalid payload"})
		return
	}

	// validate the struct
	err=utils.Validator.Struct(body)

	if err != nil {
		a.Logger.Debug("Invalid post body",zap.Error(err))

		var validationErrors validator.ValidationErrors

		if errors.As(err,&validationErrors){
			fields := make(map[string]string)

			for _, e := range validationErrors {
				fields[e.Field()] = e.Tag() 
			}
			
			utils.WriteResponse(w,http.StatusBadRequest,fields)
			return
		}

		utils.WriteResponse(w,http.StatusBadRequest,models.MessageResponse{Message: "Invalid payload"})
		return	
	}

	var deleteCode models.DeleteCode

	// get the code from the cache
	cacheDeleteKey:=fmt.Sprintf("user:%s:delete-request-code",userCtxInfo.UserId)
	err=a.Cache.GetJSON(ctx,cacheDeleteKey,&deleteCode)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			utils.WriteResponse(w, http.StatusUnauthorized, models.MessageResponse{Message: "Unauthorized"})
		} else {
			a.Logger.Error("Failed to get delete code from cache", zap.Error(err))

			utils.WriteResponse(w, http.StatusInternalServerError, models.MessageResponse{Message: "Internal server error"})
		}
		return
	}
	
	// if the codes dont match return unauthorized
	if subtle.ConstantTimeCompare([]byte(body.Code),[]byte(deleteCode.Code)) != 1 {
		a.Logger.Debug("Code does not match")

		utils.WriteResponse(w,http.StatusUnauthorized,models.MessageResponse{Message: "Unauthorized"})
		return	
	}

	err= repository.DeleteUserAccountById(ctx,a.DB,userCtxInfo.UserId)

	if err != nil {
		switch {
			case (errors.Is(err,models.ErrNotFound)) :
				a.Logger.Error("User does not exist",zap.Error(err))

				utils.WriteResponse(w,http.StatusNotFound,models.MessageResponse{Message: "Account does not exist"})
				return	
			default:
				a.Logger.Error("Error deleting the account from db",zap.Error(err))

				utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal server error"})
				return			
		}
	}
	
	err= a.Cache.Delete(ctx,cacheDeleteKey)
	if err != nil {
		a.Logger.Warn(fmt.Sprintf("Couldnt delete %s from cache",cacheDeleteKey),zap.Error(err))
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access-token",
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
		Domain: "localhost",
		Path: "/",
		Expires:  time.Unix(0, 0),
	})
	
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh-token",
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
		Domain: "localhost",
		Path: "/",
		Expires:  time.Unix(0, 0),
	})

	utils.WriteResponse(w,http.StatusOK,models.MessageResponse{Message: "Success"})
}

func (a *AuthHandler)GetMe(w http.ResponseWriter, r * http.Request){
	ctx, cancel:= context.WithTimeout(r.Context(),time.Second*10)
	defer cancel()

	// get the users details from the request context
	ctxUser,ok:= r.Context().Value(middlewares.UserCtxKey).(models.UserRequestCtx)
	if !ok {
		a.Logger.Debug("User is unauthorized")

		utils.WriteResponse(w,http.StatusUnauthorized,models.MessageResponse{Message: "Unauthorized"})
		return
	}

	// get the users details from the cache
	cacheKey:=fmt.Sprintf("user:%s",ctxUser.UserId)
	var user models.User

	err:= a.Cache.GetJSON(ctx,cacheKey,&user)

	// if the cache returned an error
	if err != nil {	
		// if it was just a cache miss
		if errors.Is(err, redis.Nil){
			a.Logger.Info("Cache Miss: User does not exist")
		}else{
		// if it was anything else
			a.Logger.Info("Cache error", zap.Error(err))
		}
	} else{
		// if the user was in the cache return that back to the user
		a.Logger.Debug("Cache hit: User info returned from cache")
		utils.WriteResponse(w,http.StatusOK,user)
		return
	}
	
	// if the user was not in the cache, get from the database
	user,err= repository.FindUserById(ctx,a.DB,ctxUser.UserId)

	if err != nil {
		if errors.Is(err, models.ErrNotFound){
			a.Logger.Debug("User does not exist",zap.Error(err))

			utils.WriteResponse(w,http.StatusNotFound,models.MessageResponse{Message: "User does not exist"})
			return
		}

		a.Logger.Error("Failed to get user info",zap.Error(err))
		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal server error"})
		return
	}

	err= a.Cache.SetJSON(ctx,cacheKey,user,time.Minute*5)
	if err != nil {
		a.Logger.Error("Failed to set user info in db",zap.Error(err))
	}

	// return the response back to the user
	utils.WriteResponse(w,http.StatusOK,user)
}