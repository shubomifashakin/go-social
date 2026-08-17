package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/shubomifashakin/go-social/internal/cache"
	"github.com/shubomifashakin/go-social/internal/middlewares"
	"github.com/shubomifashakin/go-social/internal/models"
	"github.com/shubomifashakin/go-social/pkg/utils"
	"go.uber.org/zap"
)

type PostsHandler struct {
	PostsRepo PostsRepo
	Cache    cache.CacheService
	Logger   *zap.Logger
}

type PaginatedPostResponse = models.PaginatedResponse[models.Post]

// @Summary      Create a post
// @Description  This creates a new post for the user
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        body body models.CreatePost true "Post body"
// @Success      201  {object}  models.MessageResponse
// @Failure      400  {object}  models.MessageResponse
// @Failure      401  {object}  models.MessageResponse
// @Failure      500  {object}  models.MessageResponse
// @Security     cookieAuth
// @Router       /posts [post]
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

	_,err:= p.PostsRepo.CreatePost(ctx,user.UserId,body)
	if err != nil {
		p.Logger.Error("Failed to create post in db",zap.Error(err))

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal Server Error"})
		return
	}

	// write the response 
	utils.WriteResponse(w, http.StatusCreated,models.MessageResponse{Message: "Success"})
}

// @Summary      Delete a post
// @Description  This deletes the specified post for the user
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param 		 id   path string true  "The post id"
// @Success      200  {object}  models.MessageResponse
// @Failure      400  {object}  models.MessageResponse
// @Failure      401  {object}  models.MessageResponse
// @Failure      500  {object}  models.MessageResponse
// @Security     cookieAuth
// @Router       /posts/{id} [delete]
func (p *PostsHandler) DeletePost(w http.ResponseWriter, r *http.Request){
	ctx, cancel:= context.WithTimeout(r.Context(),10 *time.Second)
	defer cancel()

	userInfo,ok:= r.Context().Value(middlewares.UserCtxKey).(models.UserRequestCtx)
	if !ok {
		p.Logger.Debug("User is unauthorized")

		utils.WriteResponse(w, http.StatusUnauthorized,models.MessageResponse{Message: "Unauthorized"})
		return
	}

	// get the post id from the request path
	postId:= r.PathValue("id")

	if postId == "" {
		p.Logger.Debug("Invalid post id")

		utils.WriteResponse(w, http.StatusBadRequest, models.MessageResponse{Message: "Invalid post id"})
		return
	}

	// delete the post from the database
	err:= p.PostsRepo.DeletePostById(ctx,postId, userInfo.UserId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound){
			utils.WriteResponse(w, http.StatusNotFound,models.MessageResponse{Message: "Post does not exist"})
			return
		}

		p.Logger.Error("Failed to delete post from DB",zap.Error(err))

		utils.WriteResponse(w, http.StatusInternalServerError,models.MessageResponse{Message: "Internal server error"})
		return
	}

	// delete the post from cache
	err=p.Cache.Delete(ctx,makePostCacheKey(postId))

	if err != nil {
		p.Logger.Error("Failed to delete post from cache",zap.Error(err))
	}

	// return the response
	utils.WriteResponse(w,http.StatusOK,models.MessageResponse{Message: "Success"})
}

// @Summary      Get a post
// @Description  This retrieves the specified post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param 		 id   path string true  "The post id"
// @Success      200  {object}  models.Post
// @Failure      400  {object}  models.MessageResponse
// @Failure      401  {object}  models.MessageResponse
// @Failure      404  {object}  models.MessageResponse
// @Failure      500  {object}  models.MessageResponse
// @Security     cookieAuth
// @Router       /posts/{id} [get]
func (p *PostsHandler) GetPost(w http.ResponseWriter, r *http.Request){
	ctx, cancel:= context.WithTimeout(r.Context(),10 *time.Second)
	defer cancel()

	// get the post id from the request path
	postId:= r.PathValue("id")

	if postId == "" {
		p.Logger.Debug("Invalid post id")

		utils.WriteResponse(w, http.StatusBadRequest, models.MessageResponse{Message: "Invalid post id"})
		return
	}
	cacheKey:=makePostCacheKey(postId)

	// get the post from the cache
	var post models.Post
	err:= p.Cache.GetJSON(ctx,cacheKey,&post)
	
	if err != nil {	
		if errors.Is(err, redis.Nil){
				p.Logger.Info("Cache Miss: Post does not exist")
		}else{
				p.Logger.Info("Cache error", zap.Error(err))
		}
	} else{
			p.Logger.Debug("Cache hit: Post returned from cache")
			utils.WriteResponse(w,http.StatusOK,post)
			return
	}

	// get the post from the database
	post,err= p.PostsRepo.GetPostById(ctx,postId)
	if err != nil {
		p.Logger.Error("Failed to get post from DB",zap.Error(err))

		utils.WriteResponse(w, http.StatusInternalServerError,models.MessageResponse{Message: "Internal server error"})
		return
	}

	// set the post in cache
	err=p.Cache.SetJSON(ctx,cacheKey,post, time.Minute*5)

	if err != nil {
		p.Logger.Error("Failed to delete post from cache",zap.Error(err))
	}

	// return the response
	utils.WriteResponse(w,http.StatusOK,post)
}

// @Summary      Get a list of posts for the logged in user
// @Description  This retrieves a list of posts for the logged in user
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        cursor  query  string false "The cursor to continue the search from"   
// @Success      200  {object}  PaginatedPostResponse
// @Failure      400  {object}  models.MessageResponse
// @Failure      401  {object}  models.MessageResponse
// @Failure      500  {object}  models.MessageResponse
// @Security     cookieAuth
// @Router       /posts [get]
func (p *PostsHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	ctx,cancel:=context.WithTimeout(r.Context(),10*time.Second)
	defer cancel()

	// get the users id from the request context
	user,ok:= r.Context().Value(middlewares.UserCtxKey).(models.UserRequestCtx)
	if !ok {
		utils.WriteResponse(w, http.StatusUnauthorized,models.MessageResponse{Message: "Unauthorized"})
		return
	}

	cursor:= r.URL.Query().Get("cursor")

	limit:=10
	posts,err:=p.PostsRepo.GetPaginatedPostsForUser(ctx,user.UserId,limit,cursor)
	if err != nil {
		p.Logger.Error("Failed to get posts",zap.Error(err))

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal server error"})
		return
	}

	var response models.PaginatedResponse[models.Post]

	if len(posts) > limit {
		response= models.PaginatedResponse[models.Post]{
			HasNextPage: len(posts) > limit,
			Data:         posts[:limit],
			Next:       posts[limit].ID,
		}	
	}else{
		response= models.PaginatedResponse[models.Post]{
			HasNextPage: false,
			Data:         posts,
			Next:       "",
		}	
	}

	// return the posts to the user
	utils.WriteResponse(w,http.StatusOK,response)
}

// @Summary      Get list of posts for a specific user
// @Description  This retrieves a list of posts for the specified user
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id   path  string  true  "The user id"
// @Param        cursor  query  string false  "The cursor to continue the search from"   
// @Success      200  {object}  PaginatedPostResponse
// @Failure      400  {object}  models.MessageResponse
// @Failure      401  {object}  models.MessageResponse
// @Failure      500  {object}  models.MessageResponse
// @Security     cookieAuth
// @Router       /users/{id}/posts [get]
func (p *PostsHandler)GetPostsForUser(w http.ResponseWriter, r *http.Request){
	ctx, cancel:= context.WithTimeout(r.Context(),10*time.Second)
	defer cancel()

	userId:= r.PathValue("id")

	if userId == ""{
		p.Logger.Debug("User id is empty")

		utils.WriteResponse(w,http.StatusBadRequest,models.MessageResponse{Message: "Invalid user id"})
		return
	}

	cursor:= r.URL.Query().Get("cursor")

	limit:=10
	posts,err:=p.PostsRepo.GetPaginatedPostsForUser(ctx,userId,limit,cursor)
	if err != nil {
		p.Logger.Error("Failed to get posts",zap.Error(err))

		utils.WriteResponse(w,http.StatusInternalServerError,models.MessageResponse{Message: "Internal server error"})
		return
	}

	var response models.PaginatedResponse[models.Post]

	if len(posts) > limit {
		response= models.PaginatedResponse[models.Post]{
			HasNextPage: len(posts) > limit,
			Data:         posts[:limit],
			Next:       posts[limit].ID,
		}	
	}else{
		response= models.PaginatedResponse[models.Post]{
			HasNextPage: false,
			Data:         posts,
			Next:       "",
		}	
	}

	// return the posts to the user
	utils.WriteResponse(w,http.StatusOK,response)
}


func makePostCacheKey(postId string) string{
	return fmt.Sprintf("post:%s",postId)
}