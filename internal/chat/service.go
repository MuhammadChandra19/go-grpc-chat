package chat

import (
	"context"
	"fmt"
	"sync"
	"time"

	v1 "github.com/MuhammadChandra19/go-grpc-chat/api/v1"
)

type Service struct {
	Repository  RepositoryInterface
	Connnection map[string]*Connection
}

type PayloadInsertUser struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type Connection struct {
	stream  v1.ChatProto_CreateStreamServer
	id      string
	roomKey string
	active  bool
	error   chan error
}

type PayloadInsertRoom struct {
	RoomKey   string     `json:"room_key"`
	Type      string     `json:"type"`
	CreatedBy string     `json:"created_by"`
	CreatedAt *time.Time `json:"created_at"`
}

type PayloadInsertUserRoom struct {
	RoomKey   string `json:"room_key"`
	UUID      string `json:"uuid"`
	UserEmail string `json:"user_email"`
}

const (
	RoomTypePrivate   = "private"
	RoomTypePublic    = "public"
	RoomTypeBroadcast = "broadcast"
)

func (s *Service) AddUserToRoom(ctx context.Context, req *v1.UserRoom) (*v1.Empty, error) {
	roomUser := UserRoom{
		RoomKey:   req.RoomKey,
		UUID:      req.UUID,
		UserEmail: req.UserEmail,
	}
	err := s.Repository.JoinRoom(ctx, roomUser)
	if err != nil {
		return nil, err
	}

	return &v1.Empty{}, nil
}

func (s *Service) CreateRoom(ctx context.Context, req *v1.Room) (*v1.Empty, error) {
	now := time.Now()

	modelRoom := Room{
		RoomKey:   req.RoomKey,
		Type:      req.Type,
		CreatedBy: req.CreatedBy,
		CreatedAt: &now,
	}

	err := s.Repository.InsertRoom(ctx, modelRoom)
	if err != nil {
		return nil, err
	}

	return &v1.Empty{}, nil
}

func (s *Service) CreateStream(connect *v1.StreamConnect, stream v1.ChatProto_CreateStreamServer) error {
	conn := &Connection{
		stream:  stream,
		id:      connect.GetName(),
		roomKey: connect.GetRoomKey(),
		active:  true,
		error:   make(chan error),
	}
	if _, ok := s.Connnection[connect.GetName()]; !ok {
		s.Connnection[connect.GetName()] = conn
	}

	return <-conn.error

}

func (s *Service) SharePoint(ctx context.Context, req *v1.Point) (*v1.Empty, error) {
	syncWait := sync.WaitGroup{}
	finish := make(chan int)

	users, err := s.Repository.GetUserInRoom(ctx, req.RoomKey)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		syncWait.Add(1)
		go func(point *v1.Point, user *UserRoom) {
			defer syncWait.Done()
			content := &v1.ResponseStream{
				IsMessage: false,
				Message:   nil,
				Point:     point,
			}
			if req.RoomKey == user.RoomKey {
				err := s.Connnection[user.UserEmail].stream.Send(content)
				if err != nil {
					fmt.Printf("Error while streaming: %v\n", err)
					s.Connnection[user.UserEmail].active = false
					s.Connnection[user.UserEmail].error <- err
				}
			}
		}(req, user)
	}

	go func() {
		syncWait.Wait()
		close(finish)
	}()

	<-finish
	return &v1.Empty{}, nil
}

func (s *Service) SendMessage(ctx context.Context, req *v1.ContentMessage) (*v1.Empty, error) {
	syncWait := sync.WaitGroup{}
	finish := make(chan int)

	users, err := s.Repository.GetUserInRoom(ctx, req.RoomKey)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		syncWait.Add(1)
		go func(messageContent *v1.ContentMessage, user *UserRoom) {
			defer syncWait.Done()
			content := &v1.ResponseStream{
				IsMessage: true,
				Message:   messageContent,
				Point:     nil,
			}
			if req.RoomKey == user.RoomKey {
				err := s.Connnection[user.UserEmail].stream.Send(content)
				if err != nil {
					fmt.Printf("Error while streaming: %v\n", err)
					s.Connnection[user.UserEmail].active = false
					s.Connnection[user.UserEmail].error <- err
				}
			}
		}(req, user)
	}

	go func() {
		syncWait.Wait()
		close(finish)
	}()

	<-finish
	return &v1.Empty{}, nil
}
