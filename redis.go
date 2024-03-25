package main

import (
	"context"
	"log"
	"net/http"

	redis "github.com/redis/go-redis/v9"
)

// NewRedisDB creates a new Redis client.
func NewRedisDB() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	return client
}

// isBlacklistedToken checks if the provided token is blacklisted in Redis.
func isBlacklistedToken(w http.ResponseWriter, r *http.Request, s *redis.Client) bool {
	// Check if token is blacklisted
	token := r.Header.Get("token")

	isBlacklisted, err := s.Exists(context.Background(), token).Result()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return true
	}

	if isBlacklisted == 1 {
		http.Error(w, "Blacklisted token", http.StatusUnauthorized)
		return true
	}
	return false
}
