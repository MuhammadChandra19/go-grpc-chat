package server

import (
	"github.com/MuhammadChandra19/go-grpc-chat/config"
	chatservice "github.com/MuhammadChandra19/go-grpc-chat/internal/chat"
	chatservicepb "github.com/MuhammadChandra19/go-grpc-chat/internal/chat/chatproto"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/storage/postgres"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
)

type Server struct{}

var conf = config.GetConfiguration()

func (as *Server) Serve() {
	var chatConnections []*chatservice.Connection
	pg := postgres.NewDatabase()
	chatRepo := chatservice.NewRepository(pg)
	s := grpc.NewServer()
	chatservicepb.RegisterChatProtoServer(s, &chatservice.Service{Connnection: chatConnections, Repository: chatRepo})
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
