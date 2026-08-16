package handlers

import (
	"context"
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
	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/mailer"
	"github.com/shubomifashakin/go-social/internal/models"
	"github.com/shubomifashakin/go-social/internal/repository"
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
	
	// FIXME: EXTRACT THIS SOMEWHERE ELSE
	var builder strings.Builder

	const welcomeTemplate = `
	<h1>Welcome to GO-Social, {{.FirstName}}!</h1>
	<p>Your account has been created successfully.</p>
	`
	tmpl := template.Must(template.New("welcome-mail").Parse(welcomeTemplate))
	err = tmpl.Execute(&builder, body)

	if err != nil {
		a.Logger.Error("Failed to generate welcome mail",zap.Error(err))
	}else {
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
		Secure: true,
		SameSite: http.SameSiteLaxMode,
		Expires:refreshExpiresAt,
		MaxAge: int(time.Until(refreshExpiresAt).Seconds()),
		Value: refreshToken,
	})

	utils.WriteResponse(w,http.StatusOK,models.MessageResponse{Message: "Success"})
}