package config

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var DBConn *pgx.Conn

func ConnectDB() *pgx.Conn {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get DATABASE_URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not found in .env")
	}

	// Connect to PostgreSQL
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	DBConn = conn
	return conn
}
