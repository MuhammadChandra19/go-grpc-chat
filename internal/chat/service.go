package chatservice

import (
	"context"
	"fmt"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/chat/chatservice"
	"sync"
	"time"
)

type service struct {
	repository  RepositoryInterface
	Connnection map[string]*Connection
}

type PayloadInsertUser struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type Connection struct {
	stream  chatservice.ChatService_CreateStreamMessageServer
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

func (s *service) RegisterUser(ctx context.Context, payload PayloadInsertUser) error {
	modelUser := User{
		Email: payload.Email,
		Name:  payload.Name,
	}

	err := s.repository.InsertUser(ctx, modelUser)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) CreateRoom(ctx context.Context, payload PayloadInsertRoom) error {
	now := time.Now()

	modelRoom := Room{
		RoomKey:   payload.RoomKey,
		Type:      payload.Type,
		CreatedBy: payload.CreatedBy,
		CreatedAt: &now,
	}

	err := s.repository.InsertRoom(ctx, modelRoom)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) CreateStreamMessage(connect *chatservice.StreamConnect, stream chatservice.ChatService_CreateStreamMessageServer) error {
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

func (s *service) SendMessage(ctx context.Context, req *chatservice.ContentMessage) (*chatservice.Empty, error) {
	syncWait := sync.WaitGroup{}
	finish := make(chan int)

	for _, conn := range s.Connnection {
		syncWait.Add(1)
		go func(messageContent *chatservice.ContentMessage, conn *Connection) {
			defer syncWait.Done()
			if conn.active {
				if req.RoomKey == conn.roomKey {
					err := conn.stream.Send(messageContent)
					fmt.Printf("Send Message to: %v\n", conn.stream)
					if err != nil {
						fmt.Printf("Error while streaming: %v\n", err)
						conn.active = false
						conn.error <- err
					}
				}
			}
		}(req, conn)
	}

	go func() {
		syncWait.Wait()
		close(finish)
	}()

	<-finish
	return &chatservice.Empty{}, nil
}
