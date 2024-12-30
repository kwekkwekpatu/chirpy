package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kwekkwekpatu/chirpy/internal/auth"
	"github.com/kwekkwekpatu/chirpy/internal/database"
	"github.com/kwekkwekpatu/chirpy/internal/util"
	_ "github.com/lib/pq"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	User_ID   uuid.UUID `json:"user_id"`
}

type ChirpSlice []Chirp

func (cfg *ApiConfig) ChirpHandler(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	util.InfoLogger.Printf("Handling chirp creation.")

	util.InfoLogger.Printf("Loading request parameter.")
	writer.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusBadRequest, "Error decoding parameters.", err)
		return
	}

	util.InfoLogger.Printf("Extracting and validating JWT token")
	tokenString, err := auth.GetBearerToken(request.Header)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "Missing authorization header.", err)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "JWT is invalid", err)
		return
	}

	util.InfoLogger.Printf("Successfully loaded chirp for user_id: %s", userID)

	util.InfoLogger.Printf("Checking if length of chirp is more than 140 characters")
	if len(params.Body) > 140 {
		util.RespondWithError(writer, request, http.StatusBadRequest, "Chirp is too long", fmt.Errorf("Chirp is too long"))
		return
	}

	util.InfoLogger.Printf("Cleaning the body of the chirp from: %s", userID)
	cleaned_body, err := cleanBody(params.Body)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusBadRequest, "Failed to clean the body!", err)
		return
	}

	chirpParams := database.CreateChirpParams{
		Body: cleaned_body,
		UserID: uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
	}

	util.InfoLogger.Printf("Attempting to create chirp with user_id: %s", userID)
	chirp, err := cfg.db.CreateChirp(request.Context(), chirpParams)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to create chirp.", err)
		return
	}

	util.InfoLogger.Printf("Generating response body from chirp.")
	responseBody := Chirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, User_ID: chirp.UserID.UUID}

	util.RespondWithJson(writer, request, http.StatusCreated, responseBody)
	util.InfoLogger.Printf("Successfully created a chirp for user: %s", userID)
	return
}

func (cfg *ApiConfig) ChirpReadHandler(writer http.ResponseWriter, request *http.Request) {
	util.InfoLogger.Printf("Handling reading of all chirps.")

	util.InfoLogger.Printf("Loading chirps from database.")
	chirpArray, err := cfg.db.ReadAllChirps(request.Context())
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to read chirps", err)
		return
	}
	util.InfoLogger.Printf("Succesfully loaded chirps.")
	var chirpSlice ChirpSlice

	util.InfoLogger.Printf("Generating response body from chirps.")
	for _, chirp := range chirpArray {
		chirpSlice = append(chirpSlice, Chirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, User_ID: chirp.UserID.UUID})
	}

	util.InfoLogger.Printf("Attempting to Marshal response.")
	util.RespondWithJson(writer, request, http.StatusOK, chirpSlice)
	return
}

func (cfg *ApiConfig) ChirpSpecificReadHandler(writer http.ResponseWriter, request *http.Request) {
	util.InfoLogger.Printf("Handling reading of chirp.")

	util.InfoLogger.Printf("Reading request ChirpID.")
	chirpIDString := request.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to read chirpID", err)
		return
	}

	dbChirp, err := cfg.db.ReadChirp(request.Context(), chirpID)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusNotFound, "chirp not found", err)
		return
	}
	responseBody := Chirp{ID: dbChirp.ID, CreatedAt: dbChirp.CreatedAt, UpdatedAt: dbChirp.UpdatedAt, Body: dbChirp.Body, User_ID: dbChirp.UserID.UUID}
	util.RespondWithJson(writer, request, http.StatusOK, responseBody)

	util.InfoLogger.Printf("Successfully read chirp.")
	return
}

func cleanBody(body string) (string, error) {
	if body == "" {
		return body, nil
	}

	unsafeWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	splitBody := strings.Split(body, " ")
	if splitBody == nil {
		return "", fmt.Errorf("Cannot split body")
	}

	for i, word := range splitBody {
		if unsafeWords[strings.ToLower(word)] {
			splitBody[i] = "****"
		}
	}

	cleanedBody := strings.Join(splitBody, " ")
	return cleanedBody, nil
}
