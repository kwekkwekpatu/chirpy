package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
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
		Body    string    `json:"body"`
		User_id uuid.UUID `json:"user_id"`
	}

	type validResponse struct {
		Cleaned_body string `json:"cleaned_body"`
	}

	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	util.InfoLogger.Printf("Handling chirp creation.")

	util.InfoLogger.Printf("Loading request parameter.")
	writer.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusBadRequest, util.InternalServerError, err)
		return
	}

	util.InfoLogger.Printf("Successfully loaded chirp for user_id: %s", params.User_id)
	if params.User_id == uuid.Nil {
		util.WarnLogger.Printf("The user_id paramater is empty. Cannot create a new chirp without a user_id.")
		util.RespondWithError(writer, request, http.StatusBadRequest, "User_ID is required", nil)
		return
	}

	util.InfoLogger.Printf("Checking if length of chirp is more than 140 characters")
	if len(params.Body) > 140 {
		util.RespondWithError(writer, request, http.StatusBadRequest, "Chirp is too long", fmt.Errorf("Chirp is too long"))
		return
	}

	util.InfoLogger.Printf("Cleaning the body of the chirp from: %s", params.User_id)
	cleaned_body, err := cleanBody(params.Body)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusBadRequest, "Failed to clean the body!", err)
		return
	}

	chirpParams := database.CreateChirpParams{
		Body: cleaned_body,
		UserID: uuid.NullUUID{
			UUID:  params.User_id,
			Valid: true,
		},
	}

	util.InfoLogger.Printf("Attempting to create chirp with user_id: %s", params.User_id)
	chirp, err := cfg.db.CreateChirp(request.Context(), chirpParams)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to create chirp.", err)
		return
	}

	util.InfoLogger.Printf("Successfully created a chirp for user: %s", params.User_id)
	util.InfoLogger.Printf("Generating response body from chirp.")
	responseBody := Chirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, User_ID: chirp.UserID.UUID}

	util.RespondWithJson(writer, request, http.StatusCreated, responseBody)
	util.InfoLogger.Printf("Successfully created chirp.")
	return
}

func (cfg *ApiConfig) ChirpReadHandler(writer http.ResponseWriter, request *http.Request) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
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
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
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
