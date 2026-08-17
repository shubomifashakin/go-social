package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shubomifashakin/go-social/internal/models"
)

type UsersRepository struct {
	Db *sql.DB
}

func (u *UsersRepository)CreateUser(ctx context.Context, user models.UserSignup) (string,error) {
	query := "INSERT INTO users(first_name, last_name, email, password, username) VALUES($1,$2,$3,$4,$5) RETURNING id"

	var id string
	err := u.Db.QueryRowContext(ctx, query, user.FirstName, user.LastName, user.Email, user.Password, user.Username).Scan(&id)
	
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

func (u *UsersRepository)DeleteUserAccountById(ctx context.Context, id string) error {
	query:= `DELETE FROM users WHERE id = $1`

	res,err:= u.Db.ExecContext(ctx,query,id)

	if err != nil {
		return err
	}
	
	rows,err:= res.RowsAffected()
	if rows < 1 {
		return models.ErrNotFound
	}

	return nil
}

func (u *UsersRepository)FindUserByUsername(ctx context.Context, username string) (models.User, error) {
	query:=`SELECT id, first_name, last_name, username, password, role, email FROM users WHERE username = $1`
	var user models.User

	if err:=u.Db.QueryRowContext(ctx,query,username).Scan(&user.ID,&user.FirstName, &user.LastName,&user.Username,&user.Password, &user.Role, &user.Email); err !=nil {
			if errors.Is(err, sql.ErrNoRows) {
				return user, models.ErrNotFound
			}
			return user, err
	}

	return user,nil
}

func (u *UsersRepository)FindUserById(ctx context.Context, userId string) (models.User, error) {
	query:=`SELECT id, first_name, last_name, username, password, role, email FROM users WHERE id = $1`
	var user models.User

	if err:=u.Db.QueryRowContext(ctx,query,userId).Scan(&user.ID,&user.FirstName, &user.LastName,&user.Username,&user.Password, &user.Role, &user.Email); err !=nil {
			if errors.Is(err, sql.ErrNoRows) {
				return user, models.ErrNotFound
			}
			return user, err
	}

	return user,nil
}

func (u *UsersRepository)CreateRefreshToken(ctx context.Context, userId string, tokenId string, expiresAt time.Time) error {
	query:= `INSERT INTO refresh_tokens(user_id, token_id, expires_at) VALUES($1, $2, $3);`

	res,err:= u.Db.ExecContext(ctx,query,userId, tokenId,expiresAt);

	if err != nil {
		return err
	}

	count,_:=res.RowsAffected()

	if count<1 {
		return errors.New("Failed to create refresh token")
	}

	return nil
}


func (u *UsersRepository)FindRefreshTokenByTokenId(ctx context.Context, tokenId string) (models.RefreshToken, error) {
	query:= `SELECT id, user_id, token_id, expires_at, created_at from refresh_tokens WHERE token_id = $1;`
	var refreshToken models.RefreshToken

	err:=u.Db.QueryRowContext(ctx, query, tokenId).Scan(&refreshToken.ID, &refreshToken.UserID, &refreshToken.TokenID, &refreshToken.ExpiresAt, &refreshToken.CreatedAt)

	if errors.Is(err,sql.ErrNoRows){
		return refreshToken, models.ErrNotFound
	}

	if err != nil {
		return refreshToken,err
	}

	return refreshToken,nil
}

func (u *UsersRepository)RotateRefreshToken(ctx context.Context, userId string, oldTokenId string, newTokenId string, expiresAt time.Time) error {
	deletePreviousTokenQuery:= `DELETE FROM refresh_tokens WHERE token_id = $1`

	insertNewTokenQuery:= `INSERT INTO refresh_tokens(user_id, token_id, expires_at) VALUES($1, $2, $3);`

	// start the transaction
	tx,err:=u.Db.BeginTx(ctx, nil)

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

func (u *UsersRepository)DeleteRefreshTokenByTokenId(ctx context.Context, tokenId string) error {
	query:= `DELETE from refresh_tokens WHERE token_id = $1;`

	res, err := u.Db.ExecContext(ctx, query, tokenId)
	if err != nil {
		return err
	}
	
	count, _ := res.RowsAffected()
	if count < 1 {
		return models.ErrNotFound
	}
	
	return nil	
}