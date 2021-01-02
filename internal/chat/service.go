package chat

import (
	"context"
	"fmt"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/chat/chatproto"
	"sync"
	"time"
)

type Service struct {
	Repository  RepositoryInterface
	Connnection []*Connection
}

type PayloadInsertUser struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type Connection struct {
	stream chatproto.ChatProto_CreateStreamMessageServer
	// stream  chatproto.ChatService_CreateStreamMessageServer
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

func (s *Service) RegisterUser(ctx context.Context, req *chatproto.User) (*chatproto.RegisterResponse, error) {
	modelUser := User{
		Email: req.Email,
		Name:  req.Name,
	}

	err := s.Repository.InsertUser(ctx, modelUser)
	if err != nil {
		return nil, err
	}

	registerRes := &chatproto.RegisterResponse{
		Result: &chatproto.User{
			Email: modelUser.Email,
			Name:  modelUser.Name,
		},
	}

	return registerRes, nil
}

func (s *Service) AddUserToRoom(ctx context.Context, req *chatproto.UserRoom) (*chatproto.Empty, error) {
	roomUser := UserRoom{
		RoomKey:   req.RoomKey,
		UUID:      req.UUID,
		UserEmail: req.UserEmail,
	}
	err := s.Repository.JoinRoom(ctx, roomUser)
	if err != nil {
		return nil, err
	}

	return &chatproto.Empty{}, nil
}

func (s *Service) CreateRoom(ctx context.Context, req *chatproto.Room) (*chatproto.Empty, error) {
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

	return &chatproto.Empty{}, nil
}

func (s *Service) CreateStreamMessage(connect *chatproto.StreamConnect, stream chatproto.ChatProto_CreateStreamMessageServer) error {
	conn := &Connection{
		stream:  stream,
		id:      connect.GetName(),
		roomKey: connect.GetRoomKey(),
		active:  true,
		error:   make(chan error),
	}
	n := 0
	for _, x := range s.Connnection {
		if x.id != connect.GetName() && x.roomKey != connect.GetRoomKey() {
			s.Connnection[n] = x
			n++
		}
	}
	s.Connnection = s.Connnection[:n]

	s.Connnection = append(s.Connnection, conn)

	return <-conn.error

}

func (s *Service) SendMessage(ctx context.Context, req *chatproto.ContentMessage) (*chatproto.Empty, error) {
	syncWait := sync.WaitGroup{}
	finish := make(chan int)

	for _, conn := range s.Connnection {
		syncWait.Add(1)
		go func(messageContent *chatproto.ContentMessage, conn *Connection) {
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
	return &chatproto.Empty{}, nil
}
