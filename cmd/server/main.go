package main

import (
	"database/sql"
	"jira-clone/internal/handler"
	"jira-clone/internal/store"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("tmpl/*")
	r.Static("/static", "./static")

	// Setup database connection (example using PostgreSQL)
	connectionString := "host=localhost port=5432 user=postgres password=soleares dbname=gira sslmode=disable"
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalln("Error connecting to the database:", err)
	}
	defer db.Close()

	// Ensure the connection is available
	err = db.Ping()
	if err != nil {
		log.Fatalln("Error pinging the database:", err)
	}

	// Create a store with the database connection
	store := store.New(db)

	// Initialize the handler with the store
	h := handler.NewHandler(store)

	h.SetupRoutes(r)
	r.Run(":8080")
}
