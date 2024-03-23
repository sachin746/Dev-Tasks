package main

import "fmt"

func main() {
	store, err := NewPostgresDB()
	redisClient := NewRedisDB()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	if err := store.initDB(); err != nil {
		fmt.Println(err)
		panic(err)
	}
	apiServer := newAPIServer(":1234", store, redisClient)
	apiServer.Run()
}
