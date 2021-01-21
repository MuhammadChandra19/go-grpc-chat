package server

import (
	"log"
	"net"
	"os"
	"os/signal"

	v1 "github.com/MuhammadChandra19/go-grpc-chat/api/v1"
	"github.com/MuhammadChandra19/go-grpc-chat/config"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/chat"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/storage/postgres"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct{}

var conf = config.GetConfiguration()

// Serve grpc
func (as *Server) Serve() {
	var chatConnections map[string]*chat.Connection
	pg := postgres.NewDatabase()

	chatRepo := chat.NewRepository(pg)
	userRepo := user.NewRepository(pg)

	s := grpc.NewServer()

	v1.RegisterChatProtoServer(s, &chat.Service{Connnection: chatConnections, Repository: chatRepo})
	v1.RegisterUserProtoServer(s, &user.Service{Repository: userRepo})

	reflection.Register(s)

	lis, err := net.Listen("tcp", ":"+conf.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		log.Println("Starting Server 1...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Wait for Control C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch
	log.Println("Stoppping the server")
	s.Stop()
	log.Println("Closing the listener")
	lis.Close()
}

func CreateAPIServer() *Server {
	return &Server{}
}
