package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClient represents the Redis client
type RedisClient struct {
	*redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient() (*RedisClient, error) {
	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6379")

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Password:     "", // no password set
		DB:           0,  // use default DB
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	log.Println("Successfully connected to Redis")
	return &RedisClient{client}, nil
}

// Close closes the Redis connection
func (rc *RedisClient) Close() error {
	return rc.Client.Close()
}

// SetJSON sets a JSON value in Redis with expiration
func (rc *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return rc.Set(ctx, key, jsonData, expiration).Err()
}

// GetJSON gets a JSON value from Redis
func (rc *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := rc.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get from Redis: %w", err)
	}

	return json.Unmarshal([]byte(data), dest)
}

// Delete removes a key from Redis
func (rc *RedisClient) Delete(ctx context.Context, key string) error {
	return rc.Del(ctx, key).Err()
}

// KeyExists checks if a key exists in Redis
func (rc *RedisClient) KeyExists(ctx context.Context, key string) (bool, error) {
	result, err := rc.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return result > 0, nil
}

// GenerateSearchCacheKey generates a cache key for flight search results (src, dest, date only)
func GenerateSearchCacheKey(source, destination, date string) string {
	return fmt.Sprintf("flight_search:%s:%s:%s", source, destination, date)
}

// GenerateSeatCacheKey generates a cache key for flight seat count
func GenerateSeatCacheKey(flightID int, date string) string {
	return fmt.Sprintf("flight_seats:%d:%s", flightID, date)
}

// GenerateBookingCacheKey generates a cache key for booking
func GenerateBookingCacheKey(bookingID int) string {
	return fmt.Sprintf("booking:%d", bookingID)
}

// GenerateTempBookingCacheKey generates a cache key for temporary booking
func GenerateTempBookingCacheKey(userID, flightID int) string {
	return fmt.Sprintf("temp_booking:%d:%d", userID, flightID)
}
