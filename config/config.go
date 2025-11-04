package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient *mongo.Client
	DBName      string
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}
	DBName = getEnv("DB_NAME", "todoapp")
}

func ConnectMongo() {
	uri := getEnv("MONGO_URI", "")
	if uri == "" {
		log.Fatal("MONGO_URI not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal("mongo connect failed:", err)
	}

	// Ping
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("mongo ping failed:", err)
	}

	MongoClient = client
	fmt.Println("Connected to MongoDB")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
