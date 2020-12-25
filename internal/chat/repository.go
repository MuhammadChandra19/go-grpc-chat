package chat

import (
	"context"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/storage"
	"log"
	"time"
)

type User struct {
	Name  string `json:"name" db:"name"`
	Email string `json:"email" db:"email"`
}

type repository struct {
	db storage.Interface
}

const (
	statementInsertUser     = `INSERT INTO "user" (name, email) values (:name,:email)`
	statementInsertRoom     = `INSERT INTO "room" (room_key, type, created_by, created_at) values (:room_key, :type, :created_by, :created_at)`
	statementInsertUserRoom = `INSERT INTO "User_room" (uuid, user_email, room_key) values (:uuid, :user_email, :room_key)`
)

// RepositoryInterface interface for using chat repo
type RepositoryInterface interface {
	// GetOne(ctx context.Context, filter map[string]interface{}) (*User, error)
	// GetAll(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*User, error)
	// Insert(ctx context.Context, userModel User) error
	// GetTotalData(ctx context.Context, filter, searchBy map[string]interface{}) int64
}

// NewRepository constructor to create chat repo
func NewRepository(data storage.Interface) RepositoryInterface {
	return &repository{
		db: data,
	}
}
