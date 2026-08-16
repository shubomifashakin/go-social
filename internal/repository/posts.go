package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

func DeletePostById(ctx context.Context, db *sql.DB, postId string, userId string) error {
	query:= `DELETE FROM posts WHERE id= $1 AND user_id = $2`

	res,err:=db.ExecContext(ctx,query,postId,userId)

	if err != nil {
		return err
	}

	count, _ := res.RowsAffected()
	if count < 1 {
		return models.ErrNotFound
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

func GetPaginatedPostsForUser(ctx context.Context, db *sql.DB, userId string, limit int, cursor string) ([]models.Post, error){
	posts:=[]models.Post{}
	limit=limit+1

	if cursor != "" {
		var createdAt time.Time
		createdAtOfLastPost:= `SELECT created_at FROM posts WHERE id = $1 AND user_id= $2`
		err:=db.QueryRowContext(ctx,createdAtOfLastPost,cursor, userId).Scan(&createdAt)
		if err != nil {
			if errors.Is(err,sql.ErrNoRows){
				return posts,models.ErrNotFound
			}
		}

		query:= `SELECT id, user_id, content, created_at, updated_at FROM posts WHERE user_id = $1 AND created_at < $2 ORDER BY created_at DESC LIMIT $3`

		rows,err:=db.QueryContext(ctx,query,userId,createdAt,limit)	

		if err != nil {
			return posts,err
		}
		defer rows.Close()
		
		for rows.Next() {
			post:= models.Post{}
			err:=rows.Scan(&post.ID, &post.UserID, &post.Content, &post.CreatedAt, &post.UpdatedAt)
			if err != nil {
				return posts,err
			}
	
			posts=append(posts, post)
		}
	
		if rows.Err() != nil{
			return posts,rows.Err()
		}

		return posts,nil
	}

	// if theres no cursor
	query:= `SELECT id, user_id, content, created_at, updated_at FROM posts WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`

	rows,err:=db.QueryContext(ctx,query,userId,limit)
	if err != nil {
		return posts,err
	}
	defer rows.Close()
	
	for rows.Next() {
		post:= models.Post{}
		err:=rows.Scan(&post.ID, &post.UserID, &post.Content, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return posts,err
		}

		posts=append(posts, post)
	}

	if rows.Err() != nil{
		return posts,rows.Err()
	}

	return posts,nil
}