package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shubomifashakin/go-social/internal/models"
)

func CreateUser(ctx context.Context, db *sql.DB, user models.UserSignup) (string,error) {
	query := "INSERT INTO users(first_name, last_name, email, password, username) VALUES($1,$2,$3,$4,$5) RETURNING id"

	var id string
	err := db.QueryRowContext(ctx, query, user.FirstName, user.LastName, user.Email, user.Password, user.Username).Scan(&id)
	
	if err != nil {
		var pgErr *pgconn.PgError
		
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return "", models.ErrDuplicateEntry
			case "23502":
				return "", models.ErrMissingField
			}
		}
		
		return "", err
	}

	return id,nil
}

func DeleteUserAccountById(ctx context.Context, db *sql.DB, id string) error {
	query:= `DELETE FROM users WHERE id = $1`

	res,err:= db.ExecContext(ctx,query,id)

	if err != nil {
		return err
	}
	
	rows,err:= res.RowsAffected()
	if rows < 1 {
		return models.ErrNotFound
	}

	return nil
}

func FindUserByUsername(ctx context.Context, db *sql.DB, username string) (models.User, error) {
	query:=`SELECT id, first_name, last_name, username, password, role, email FROM users WHERE username = $1`
	var user models.User

	if err:=db.QueryRowContext(ctx,query,username).Scan(&user.ID,&user.FirstName, &user.LastName,&user.Username,&user.Password, &user.Role, &user.Email); err !=nil {
			if errors.Is(err, sql.ErrNoRows) {
				return user, models.ErrNotFound
			}
			return user, err
	}

	return user,nil
}

func FindUserById(ctx context.Context, db *sql.DB, userId string) (models.User, error) {
	query:=`SELECT id, first_name, last_name, username, password, role, email FROM users WHERE id = $1`
	var user models.User

	if err:=db.QueryRowContext(ctx,query,userId).Scan(&user.ID,&user.FirstName, &user.LastName,&user.Username,&user.Password, &user.Role, &user.Email); err !=nil {
			if errors.Is(err, sql.ErrNoRows) {
				return user, models.ErrNotFound
			}
			return user, err
	}

	return user,nil
}

func CreateRefreshToken(ctx context.Context, db *sql.DB, userId string, tokenId string, expiresAt time.Time) error {
	query:= `INSERT INTO refresh_tokens(user_id, token_id, expires_at) VALUES($1, $2, $3);`

	res,err:= db.ExecContext(ctx,query,userId, tokenId,expiresAt);

	if err != nil {
		return err
	}

	count,_:=res.RowsAffected()

	if count<1 {
		return errors.New("Failed to create refresh token")
	}

	return nil
}


func FindRefreshTokenByTokenId(ctx context.Context, db *sql.DB, tokenId string) (models.RefreshToken, error) {
	query:= `SELECT id, user_id, token_id, expires_at, created_at from refresh_tokens WHERE token_id = $1;`
	var refreshToken models.RefreshToken

	err:=db.QueryRowContext(ctx, query, tokenId).Scan(&refreshToken.ID, &refreshToken.UserID, &refreshToken.TokenID, &refreshToken.ExpiresAt, &refreshToken.CreatedAt)

	if errors.Is(err,sql.ErrNoRows){
		return refreshToken, models.ErrNotFound
	}

	if err != nil {
		return refreshToken,err
	}

	return refreshToken,nil
}

func RotateRefreshToken(ctx context.Context, db *sql.DB, userId string, oldTokenId string, newTokenId string, expiresAt time.Time) error {
	deletePreviousTokenQuery:= `DELETE FROM refresh_tokens WHERE token_id = $1`

	insertNewTokenQuery:= `INSERT INTO refresh_tokens(user_id, token_id, expires_at) VALUES($1, $2, $3);`

	// start the transaction
	tx,err:=db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}
	defer tx.Rollback()

	// delete the old token
	_,err=tx.ExecContext(ctx,deletePreviousTokenQuery,oldTokenId)

	if err != nil {
		return err
	}

	// insert the new token
	_,err= tx.ExecContext(ctx,insertNewTokenQuery,userId,newTokenId,expiresAt)

	if err != nil {
		return err
	}

	// commit the transaction
	err=tx.Commit()

	if err != nil {
		return err
	}
	
	return nil
}

func DeleteRefreshTokenByTokenId(ctx context.Context,db *sql.DB, tokenId string) error {
	query:= `DELETE from refresh_tokens WHERE token_id = $1;`

	res, err := db.ExecContext(ctx, query, tokenId)
	if err != nil {
		return err
	}
	
	count, _ := res.RowsAffected()
	if count < 1 {
		return models.ErrNotFound
	}
	
	return nil	
}