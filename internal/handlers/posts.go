package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/middlewares"
	"github.com/shubomifashakin/go-social/internal/models"
	"github.com/shubomifashakin/go-social/internal/repository"
	"github.com/shubomifashakin/go-social/pkg/utils"
	"go.uber.org/zap"
)

type PostsHandler struct {
	DB       *sql.DB
	Cache    *cache.Cache
	Logger   *zap.Logger
}

func (p *PostsHandler) CreatePost(w http.ResponseWriter, r *http.Request){
	ctx,cancel:= context.WithTimeout(r.Context(),10*time.Second)
	defer cancel()

	// get the user info from the request and if not present return unauthorized
	user,ok:= r.Context().Value(middlewares.UserCtxKey).(models.UserRequestCtx)

	if !ok {
		p.Logger.Debug("User info is not in request context")

		utils.WriteResponse(w,http.StatusUnauthorized,models.MessageResponse{Message: "Unauthorized"})
		return
	}
	
	var body models.CreatePost

	// extract the post body and if extraction failed return 400
	if err:=json.NewDecoder(r.Body).Decode(&body); err != nil {
		p.Logger.Debug("Invalid payload",zap.Error(err))

		utils.WriteResponse(w,http.StatusBadRequest,models.MessageResponse{Message: "Invalid payload"})
		return
	}
	
	// validate the post body and if invalid, return 400
	if err:=utils.Validator.Struct(body); err != nil {
		p.Logger.Debug("Invalid payload",zap.Error(err))
		
		var validationErrors validator.ValidationErrors
		
		if errors.As(err, &validationErrors) {
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

	_,err:= repository.CreatePost(ctx,p.DB,user.UserId,body)
	if err != nil {
		p.Logger.Error("Failed to create post in db",zap.Error(err))

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal Server Error"})
		return
	}

	// write the response 
	utils.WriteResponse(w, http.StatusCreated,models.MessageResponse{Message: "Success"})
}