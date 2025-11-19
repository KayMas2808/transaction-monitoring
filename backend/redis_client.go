package main

import (
	"context"

	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var ctx = context.Background()

func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
	} else {
		log.Println("Connected to Redis successfully")
	}
}

func IncrementVelocity(userID string) (int64, error) {
	key := fmt.Sprintf("velocity:%s", userID)
	count, err := redisClient.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if count == 1 {
		redisClient.Expire(ctx, key, time.Minute)
	}

	return count, nil
}

func AddTransactionAmount(userID string, amount float64) error {
	key := fmt.Sprintf("history:%s", userID)
	err := redisClient.LPush(ctx, key, amount).Err()
	if err != nil {
		return err
	}
	redisClient.LTrim(ctx, key, 0, 49)
	redisClient.Expire(ctx, key, 24*time.Hour)
	return nil
}

func GetRecentAmounts(userID string) ([]float64, error) {
	key := fmt.Sprintf("history:%s", userID)
	vals, err := redisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var amounts []float64
	for _, v := range vals {
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			amounts = append(amounts, f)
		}
	}
	return amounts, nil
}
