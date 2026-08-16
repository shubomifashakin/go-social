package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/shubomifashakin/go-social/internal/models"
)

func CreatePost(ctx context.Context,db *sql.DB,userId string, post models.CreatePost) (string,error){
	query:= `INSERT INTO posts(user_id, content) VALUES ($1,$2) RETURNING id`

	var id string

	err:=db.QueryRowContext(ctx,query,userId,post.Content).Scan(&id)

	if err != nil {
		return "",err
	}

	return id, nil
}

func DeletePostById(ctx context.Context, db *sql.DB, postId string) error {
	query:= `DELETE FROM posts WHERE id= $1`

	_,err:=db.ExecContext(ctx,query,postId)

	if err != nil {
		return err
	}

	return nil
}
func GetPostById(ctx context.Context, db *sql.DB, postId string) (models.Post, error) {
	query:= `SELECT id, user_id, content, created_at FROM posts WHERE id= $1`

	var post models.Post
	err:=db.QueryRowContext(ctx,query,postId).Scan(&post.ID,&post.UserID,&post.Content, &post.CreatedAt)

	if err != nil {
		if errors.Is(err,sql.ErrNoRows){
			return post,models.ErrNotFound
		}
		
		return post, err
	}

	return post,nil
}