package user

import (
	"context"
	"log"

	"github.com/MuhammadChandra19/go-grpc-chat/internal/errors"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/storage"
)

type User struct {
	Username string `json:"username" db:"username"`
	Name     string `json:"name" db:"name"`
	Email    string `json:"email" db:"email"`
	PhotoURL string `json:"photoUrl" db:"photo_url"`
}

type repository struct {
	db storage.Interface
}

const (
	queryUser               = `SELECT email, username, name, photo_url FROM "user"`
	statementInsertUser     = `INSERT INTO "user" (name, email, photo_url, username) values (:name,:email,:photo_url,:username)`
	statementGetUserByEmail = `SELECT * FROM "user" where email = :email`
	statementSearchUser     = `SELECT * FROM "user" where username like :username or name like :name`
)

var (
	// ErrDataNotFound error data tidak ditemukan
	ErrDataNotFound = errors.N(errors.CodeNotFoundError, "no data found")
)

type RepositoryInterface interface {
	InsertUser(ctx context.Context, userModel User) error
	getByEmail(ctx context.Context, email string) (*User, error)
	getUserList(ctx context.Context, query string) ([]*User, error)
	GetOne(ctx context.Context, filter map[string]interface{}) (*User, error)
}

func (r *repository) InsertUser(ctx context.Context, userModel User) error {
	err := r.db.Exec(ctx, statementInsertUser, userModel)
	if err != nil {
		log.Println("Error: Insert User, ", err)
		return err
	}
	return nil
}

func (r *repository) GetOne(ctx context.Context, filter map[string]interface{}) (*User, error) {
	queryParams := r.db.GenerateQueryParams(queryUser, filter, nil)
	response := User{}
	err := r.db.Query(ctx, queryParams, filter, &response, false)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (r *repository) getByEmail(ctx context.Context, email string) (*User, error) {
	var response *User
	filter := map[string]interface{}{
		"email": email,
	}
	queryParams := r.db.GenerateQueryParams(statementGetUserByEmail, filter, nil)
	err := r.db.Query(ctx, queryParams, filter, &response, false)
	if err != nil {
		return nil, err
	}
	if response.Email == "" {
		return nil, ErrDataNotFound
	}

	return response, nil
}

func (r *repository) getUserList(ctx context.Context, query string) ([]*User, error) {
	var response []*User
	filter := map[string]interface{}{
		"username": query,
		"name":     query,
	}
	queryParams := r.db.GenerateQueryParams(queryUser, nil, filter)

	err := r.db.Query(ctx, queryParams, filter, &response, false)
	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, ErrUserNotFound
	}

	return response, nil

}

func NewRepository(data storage.Interface) RepositoryInterface {
	return &repository{
		db: data,
	}

}
