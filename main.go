package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kwekkwekpatu/chirpy/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	mu             sync.Mutex
	fileserverHits int
	db             *database.Queries
	platform       string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	User_ID   uuid.UUID `json:"user_id"`
}

type ChirpSlice []Chirp

var internalServerErrorJSON = []byte(`{"error":"Something went wrong"}`)

var (
	InfoLogger  *log.Logger
	WarnLogger  *log.Logger
	ErrorLogger *log.Logger
)

func initLoggers() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarnLogger = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	initLoggers()

	InfoLogger.Printf("Loading environment variables.")
	err := godotenv.Load("/app/.env")
	if err != nil {
		ErrorLogger.Println(err)
		return
	}
	platform := os.Getenv("PLATFORM")
	dbURL := os.Getenv("DB_URL")
	InfoLogger.Printf("Succesfully loaded environment variables.")
	InfoLogger.Printf("Succesfully DB_URL: %s", dbURL)

	InfoLogger.Printf("Loading Postgres database.")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		ErrorLogger.Println(err)
	}
	dbQueries := database.New(db)
	fmt.Println(dbQueries)
	InfoLogger.Printf("Succesfully loaded database.")

	mux := http.NewServeMux()
	apiCfg := apiConfig{db: dbQueries, platform: platform}

	port := os.Getenv("PORT")
	port = ":" + port

	mux.Handle("GET /app/*", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.middlewareMetricsResult)
	mux.HandleFunc("GET /api/chirps", apiCfg.chirpReadHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.chirpSpecificReadHandler)
	mux.HandleFunc("POST /api/reset", apiCfg.middlewareMetricsReset)
	mux.HandleFunc("POST /api/chirps", apiCfg.chirpHandler)
	mux.HandleFunc("POST /api/users", apiCfg.userHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.adminReset)

	mux.HandleFunc("GET /", dockerHandler)

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}
	fmt.Println(server.ListenAndServe())
}

func dockerHandler(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.NotFound(writer, request)
		return
	}
	filePath := "/public/index.html"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("File not found: %s", filePath)
		http.Error(writer, "File not found", http.StatusNotFound)
		return
	}
	http.ServeFile(writer, request, filePath)
}

func readinessHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.mu.Lock()
		defer cfg.mu.Unlock()
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMetricsResult(writer http.ResponseWriter, request *http.Request) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	body := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits)
	writer.WriteHeader(200)
	writer.Header().Add("Content-Type", "text/html")
	writer.Write([]byte(body))
}

func (cfg *apiConfig) middlewareMetricsReset(writer http.ResponseWriter, request *http.Request) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	cfg.fileserverHits = 0
	writer.WriteHeader(200)
}

func (cfg *apiConfig) chirpHandler(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Body    string    `json:"body"`
		User_id uuid.UUID `json:"user_id"`
	}

	type validResponse struct {
		Cleaned_body string `json:"cleaned_body"`
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	InfoLogger.Printf("Handling chirp creation.")

	InfoLogger.Printf("Loading request parameter.")
	writer.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		writer.WriteHeader(500)
		writer.Write(internalServerErrorJSON)
		return
	}

	InfoLogger.Printf("Successfully loaded chirp for user_id: %s", params.User_id)
	if params.User_id == uuid.Nil {
		WarnLogger.Printf("The user_id paramater is empty. Cannot create a new chirp without a user_id.")
		writer.WriteHeader(400)
		writer.Write([]byte(`{"error":"User_ID is required"}`))
		return
	}

	InfoLogger.Printf("Checking if length of chirp is more than 140 characters")
	if len(params.Body) > 140 {
		ErrorLogger.Print("Chirp is too long.")
		response := errorResponse{Error: "Chirp is too long"}
		dat, err := json.Marshal(response)
		if err != nil {
			writer.WriteHeader(500)
			writer.Write(internalServerErrorJSON)
			return
		}
		writer.WriteHeader(400)
		writer.Write(dat)
		return
	}

	InfoLogger.Printf("Cleaning the body of the chirp from: %s", params.User_id)
	cleaned_body, err := cleanBody(params.Body)
	if err != nil {
		ErrorLogger.Print("Failed to clean the body!")
		writer.WriteHeader(500)
		writer.Write(internalServerErrorJSON)
		return
	}

	chirpParams := database.CreateChirpParams{
		Body: cleaned_body,
		UserID: uuid.NullUUID{
			UUID:  params.User_id,
			Valid: true,
		},
	}

	InfoLogger.Printf("Attempting to create chirp with user_id: %s", params.User_id)
	chirp, err := cfg.db.CreateChirp(request.Context(), chirpParams)
	if err != nil {
		ErrorLogger.Println(err)
		writer.WriteHeader(500)
		writer.Write([]byte(`{"error":"Failed to create chirp."}`))
		return
	}

	InfoLogger.Printf("Successfully created a chirp for user: %s", params.User_id)
	InfoLogger.Printf("Generating response body from chirp.")
	responseBody := Chirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, User_ID: chirp.UserID.UUID}

	dat, err := json.Marshal(responseBody)
	if err != nil {
		ErrorLogger.Println(err)
		writer.WriteHeader(500)
		writer.Write([]byte(`{"error":"Failed to Marshal response."}`))
		return
	}
	writer.WriteHeader(201)
	writer.Write(dat)
	InfoLogger.Printf("Successfully created chirp.")
	return
}

func (cfg *apiConfig) chirpReadHandler(writer http.ResponseWriter, request *http.Request) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	InfoLogger.Printf("Handling reading of all chirps.")

	InfoLogger.Printf("Loading chirps from database.")
	chirpArray, err := cfg.db.ReadAllChirps(request.Context())
	if err != nil {
		ErrorLogger.Println(err)
		writer.WriteHeader(500)
		writer.Write([]byte(`{"error":"Failed to read chirps"`))
	}
	InfoLogger.Printf("Succesfully loaded chirps.")
	var chirpSlice ChirpSlice

	InfoLogger.Printf("Generating response body from chirps.")
	for _, chirp := range chirpArray {
		chirpSlice = append(chirpSlice, Chirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, User_ID: chirp.UserID.UUID})
	}

	InfoLogger.Printf("Attempting to Marshal response.")
	dat, err := json.Marshal(chirpSlice)
	if err != nil {
		ErrorLogger.Println(err)
		writer.WriteHeader(500)
		writer.Write(internalServerErrorJSON)
		return
	}
	writer.WriteHeader(200)
	writer.Write(dat)
	return
}

func (cfg *apiConfig) chirpSpecificReadHandler(writer http.ResponseWriter, request *http.Request) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	InfoLogger.Printf("Handling reading of chirp.")

	InfoLogger.Printf("Reading request ChirpID.")
	chirpIDString := request.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		ErrorLogger.Println(err)
		writer.WriteHeader(500)
		writer.Write([]byte(`{"error":"Failed to read chirps"`))
	}

	dbChirp, err := cfg.db.ReadChirp(request.Context(), chirpID)
	if err != nil {
		ErrorLogger.Println(err)
		writer.WriteHeader(404)
		writer.Write([]byte(`{"error":"chirp not found"`))
	}
	responseBody := Chirp{ID: dbChirp.ID, CreatedAt: dbChirp.CreatedAt, UpdatedAt: dbChirp.UpdatedAt, Body: dbChirp.Body, User_ID: dbChirp.UserID.UUID}
	dat, err := json.Marshal(responseBody)
	if err != nil {
		ErrorLogger.Println(err)
		writer.WriteHeader(500)
		writer.Write([]byte(`{"error":"Failed to Marshal response."}`))
		return
	}
	writer.WriteHeader(200)
	writer.Write(dat)
	InfoLogger.Printf("Successfully read chirp.")
	return
}

func (cfg *apiConfig) userHandler(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	InfoLogger.Printf("Handling user creation.")

	InfoLogger.Printf("Loading request parameter.")
	writer.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		ErrorLogger.Println(err)
		writer.WriteHeader(500)
		writer.Write(internalServerErrorJSON)
		return
	}

	InfoLogger.Printf("Successfully loaded email: %s", params.Email)
	if params.Email == "" {
		WarnLogger.Printf("The email paramater is empty. Cannot create a new user without an email.")
		writer.WriteHeader(400)
		writer.Write([]byte(`{"error":"Email is required"}`))
		return
	}

	InfoLogger.Printf("Attempting to create user with email: %s", params.Email)
	user, err := cfg.db.CreateUser(request.Context(), params.Email)
	if err != nil {
		ErrorLogger.Println(err)
		writer.WriteHeader(500)
		writer.Write([]byte(`{"error":"Failed to create user."}`))
		return
	}

	InfoLogger.Printf("Successfully created a user for email: %s", params.Email)
	InfoLogger.Printf("Generating response body from user.")
	responseBody := User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email}
	dat, err := json.Marshal(responseBody)
	if err != nil {
		ErrorLogger.Println(err)
		writer.WriteHeader(500)
		writer.Write([]byte(`{"error":"Failed to Marshal response."}`))
		return
	}
	writer.WriteHeader(201)
	writer.Write(dat)
	InfoLogger.Printf("Successfully created user.")
	return
}

func (cfg *apiConfig) adminReset(writer http.ResponseWriter, request *http.Request) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if cfg.platform != "dev" {
		writer.WriteHeader(403)
		return
	}

	err := cfg.db.DeleteUsers(request.Context())
	if err != nil {
		writer.WriteHeader(500)
		writer.Write([]byte(`{"error":"Failed to delete users."}`))
		return
	}

	writer.WriteHeader(200)
	return
}

func cleanBody(body string) (string, error) {
	if body == "" {
		return body, nil
	}

	unsafe_words := []string{"kerfuffle", "sharbert", "fornax"}
	split_body := strings.Split(body, " ")
	if split_body == nil {
		return "", fmt.Errorf("Cannot split body")
	}

	for i, value := range split_body {
		for _, unsafe_word := range unsafe_words {
			if strings.ToLower(value) == unsafe_word {
				split_body[i] = "****"
			}
		}
	}
	cleaned_body := strings.Join(split_body, " ")
	return cleaned_body, nil
}
