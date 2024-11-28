package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
)

func main() {
	// Set up the Redis cluster client
	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			"127.0.0.1:7001", // Replace with your Redis cluster node addresses
			"127.0.0.1:7002",
			"127.0.0.1:7003",
			"127.0.0.1:7004",
			"127.0.0.1:7005",
			"127.0.0.1:7006",
		},
		Password: "", // Set if password is required for your cluster
	})

	// Create a context to manage the requests
	ctx := context.Background()

	// Test the connection
	_, err := clusterClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis cluster: %v", err)
	}
	fmt.Println("Successfully connected to Redis cluster!")

	// Set a key in the cluster
	err = clusterClient.Set(ctx, "example_key", "hello, redis cluster!", 0).Err()
	if err != nil {
		log.Fatalf("Error setting key: %v", err)
	}
	fmt.Println("Key set successfully!")

	// Get the key from the cluster
	val, err := clusterClient.Get(ctx, "example_key").Result()
	if err != nil {
		log.Fatalf("Error getting key: %v", err)
	}
	fmt.Printf("Got value: %s\n", val)

	// Close the cluster client when done
	defer clusterClient.Close()
}
