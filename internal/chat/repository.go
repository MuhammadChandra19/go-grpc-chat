package chat

import (
	"context"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/errors"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/storage"
	"log"
	"time"
)

type User struct {
	Name  string `json:"name" db:"name"`
	Email string `json:"email" db:"email"`
}

type Room struct {
	RoomKey   string     `db:"room_key"`
	Type      string     `db:"type"`
	CreatedBy string     `db:"created_by"`
	CreatedAt *time.Time `db:"created_at"`
}

type UserRoom struct {
	UUID      string `db:"uuid"`
	UserEmail string `db:"user_email"`
	RoomKey   string `db:"room_key"`
}

type repository struct {
	db storage.Interface
}

const (
	statementInsertUser    = `INSERT INTO "user" (name, email) values (:name,:email)`
	statementInsertRoom    = `INSERT INTO "room" (room_key, type, created_by, created_at) values (:room_key, :type, :created_by, :created_at)`
	statementUserJoinRoom  = `INSERT INTO "user_room" (uuid, user_email, room_key) values (:uuid, :user_email, :room_key)`
	statementGetUserInRoom = `SELECT * from "user_room" WHERE room_key = :room_key`
)

var (
	// ErrDataNotFound error data tidak ditemukan
	ErrDataNotFound = errors.N(errors.CodeNotFoundError, "no data found")
)

// RepositoryInterface interface for using chat repo
type RepositoryInterface interface {
	InsertUser(ctx context.Context, userModel User) error
	InsertRoom(ctx context.Context, roomModel Room) error
	JoinRoom(ctx context.Context, userRoomModel UserRoom) error
	GetUserInRoom(ctx context.Context, roomKey string) ([]*UserRoom, error)
}

func (r *repository) InsertUser(ctx context.Context, userModel User) error {
	err := r.db.Exec(ctx, statementInsertUser, userModel)
	if err != nil {
		log.Println("Error: Insert User, ", err)
		return err
	}
	return nil
}

func (r *repository) InsertRoom(ctx context.Context, roomModel Room) error {
	err := r.db.Exec(ctx, statementInsertRoom, roomModel)
	if err != nil {
		log.Println("Error: Insert User, ", err)
		return err
	}
	return nil
}

func (r *repository) JoinRoom(ctx context.Context, userRoomModel UserRoom) error {
	err := r.db.Exec(ctx, statementUserJoinRoom, userRoomModel)
	if err != nil {
		log.Println("Error: Insert User, ", err)
		return err
	}
	return nil
}

func (r *repository) GetUserInRoom(ctx context.Context, roomKey string) ([]*UserRoom, error) {
	var response []*UserRoom
	filter := map[string]interface{}{
		"room_key": roomKey,
	}
	queryParams := r.db.GenerateQueryParams(statementGetUserInRoom, filter, nil)
	err := r.db.Query(ctx, queryParams, filter, &response, false)
	if err != nil {
		return nil, err
	}
	if len(response) == 0 {
		return nil, ErrDataNotFound
	}

	return response, nil
}

// NewRepository constructor to create chat repo
func NewRepository(data storage.Interface) RepositoryInterface {
	return &repository{
		db: data,
	}
}
