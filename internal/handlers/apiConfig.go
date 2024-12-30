package handlers

import (
	"database/sql"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/kwekkwekpatu/chirpy/internal/database"
	"github.com/kwekkwekpatu/chirpy/internal/util"
	_ "github.com/lib/pq"
)

type ApiConfig struct {
	mu             sync.Mutex
	fileserverHits int
	db             *database.Queries
	platform       string
	jwtSecret      string
}

var APIConfig *ApiConfig

func init() {
	util.InfoLogger.Printf("Loading environment variables.")
	err := godotenv.Load("/app/.env")
	if err != nil {
		util.ErrorLogger.Println(err)
		return
	}
	platform := os.Getenv("PLATFORM")
	dbURL := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	util.InfoLogger.Printf("Succesfully loaded environment variables.")

	util.InfoLogger.Printf("Loading Postgres database.")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		util.ErrorLogger.Println(err)
	}
	dbQueries := database.New(db)
	util.InfoLogger.Printf("Succesfully loaded database.")

	APIConfig = &ApiConfig{db: dbQueries, platform: platform, jwtSecret: jwtSecret}
}
