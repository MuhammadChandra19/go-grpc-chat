package main

import (
	"github.com/MuhammadChandra19/go-grpc-chat/internal/storage/postgres"
)

func main() {
	// Start Migration
	db := postgres.NewDatabase()
	err := db.Migrate(false)
	if err != nil {
		// if you cant connect to db why bother
		panic(err)
	}

	// Serve HTTP Server
	// api := server.CreateAPIServer()
	// api.Serve()
}
