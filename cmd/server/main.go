package main

import (
	"database/sql"
	"jira-clone/internal/handler"
	"jira-clone/internal/store"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("tmpl/*")
	r.Static("/static", "./static")
	localLaunch := 1
	connectionString := "host=localhost port=5432 user=postgres password=soleares dbname=gira sslmode=disable"
	// Setup database connection (example using PostgreSQL)
	if localLaunch == 0 {

		dbName := os.Getenv("PGDATABASE")
		pgHost := os.Getenv("PGHOST")
		pgPassword := os.Getenv("PGPASSWORD")
		pgPort := os.Getenv("PGPORT")
		pgUser := os.Getenv("PGUSER")

		connectionString = "host=" + pgHost + " port=" + pgPort + " user=" + pgUser + " password=" + pgPassword + " dbname=" + dbName + " sslmode=disable"
	}
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
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	h.SetupRoutes(r)
	r.Run("0.0.0.0:" + port)
}
