package main

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/parthpati1102/todo-gin-jwt/config"
	"github.com/parthpati1102/todo-gin-jwt/routes"
)

func uuidNewInternal() string {
	return uuid.NewString()
}

func main() {
	// load envs and connect mongo
	config.LoadEnv()
	config.ConnectMongo()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := routes.SetupRouter()
	addr := fmt.Sprintf(":%s", port)
	println("Server running at http://localhost" + addr)
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
