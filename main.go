package main

import (
	"fmt"
	_ "github.com/swaggo/http-swagger"
)

// @title Dev-Tasks
// @version 1.0
// @description Blog app for adding learn materials.
// @host localhost:1234
// @BasePath /
func main() {
	// Initialize database connections
	store, err := NewPostgresDB()
	redisClient := NewRedisDB()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// Ensure database schema is initialized
	if err := store.InitDB(); err != nil {
		fmt.Println(err)
		panic(err)
	}

	// Initialize API server and start listening for requests
	apiServer := newAPIServer(":1234", store, redisClient)
	apiServer.Run()
}
