package redis_utils

import (
	"fmt"
	"log"
	"os"
	"github.com/redis/go-redis/v9"
	"context"
)


func InitRedis(clear bool) (*redis.Client, context.Context) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_CONNECTION"), // Redis connection string
		Password: "",                            // No password set
		DB:       0,                             // Use default DB
	})

	if clear {
		removeAllHashKeys(redisClient, "*")
	}

	redisCtx := context.Background()

	return redisClient, redisCtx
}

func InitRedisCluster(clear bool) (*redis.ClusterClient, context.Context) {
	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{":7001", ":7002", ":7003", ":7004", ":7005", ":7006"},
		Password: "", // Set if password is required for your cluster
	})

	if clear {
		removeAllHashKeysCluster(clusterClient, "*")
	}

	redisCtx := context.Background()

	return clusterClient, redisCtx
}


// Remove all hash keys
func removeAllHashKeys(redisClient *redis.Client, pattern string) error {

	redisCtx := context.Background()
	// Find keys matching the pattern
	keys, err := redisClient.Keys(redisCtx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to fetch keys: %v", err)
	}

	if len(keys) == 0 {
		log.Println("No keys match the given pattern")
		return nil
	}

	// Delete the keys
	deletedCount, err := redisClient.Del(redisCtx, keys...).Result()
	if err != nil {
		return fmt.Errorf("failed to delete keys: %v", err)
	}

	log.Printf("Deleted %d keys matching the pattern '%s'", deletedCount, pattern)
	return nil
}


// Remove all hash keys
func removeAllHashKeysCluster(redisClient *redis.ClusterClient, pattern string) error {

	redisCtx := context.Background()
	// Find keys matching the pattern
	keys, err := redisClient.Keys(redisCtx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to fetch keys: %v", err)
	}

	if len(keys) == 0 {
		log.Println("No keys match the given pattern")
		return nil
	}

	// Delete the keys
	deletedCount, err := redisClient.Del(redisCtx, keys...).Result()
	if err != nil {
		return fmt.Errorf("failed to delete keys: %v", err)
	}

	log.Printf("Deleted %d keys matching the pattern '%s'", deletedCount, pattern)
	return nil
}