package main

import (
	"log"

	"goquickactu/db"
	"goquickactu/routes"
)

func main() {
	if _, err := db.Connect(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	r := routes.SetupRouter()
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}