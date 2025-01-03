package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/kwekkwekpatu/chirpy/internal/auth"
	"github.com/kwekkwekpatu/chirpy/internal/util"
	_ "github.com/lib/pq"
)

func (cfg *ApiConfig) UpgradeUser(writer http.ResponseWriter, request *http.Request) {
	type data struct {
		UserID uuid.UUID `json:"user_id"`
	}
	type parameters struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}

	const validEvent = "user.upgraded"

	util.InfoLogger.Printf("Handling user upgrade webhook.")

	util.InfoLogger.Printf("Loading request parameter.")
	writer.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, util.InternalServerError, err)
		return
	}

	util.InfoLogger.Printf("Successfully loaded event: %s", params.Event)
	if params.Event != validEvent {
		util.InfoLogger.Printf("This type of event will not be handled")
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	util.InfoLogger.Printf("Checking for apiKey")
	apiKey, err := auth.GetAPIKey(request.Header)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "Failed to load apiKey", err)
		return
	}

	if apiKey != cfg.polkaKey {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "Invalid apiKey", err)
		return
	}

	util.InfoLogger.Printf("Processing upgrade")

	util.InfoLogger.Printf("Checking if user exists")
	_, err = cfg.db.GetUserByID(request.Context(), params.Data.UserID)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusNotFound, "Failed to find user", err)
		return
	}

	util.InfoLogger.Printf("Attempting to upgrade user with id: %s", params.Data.UserID)
	err = cfg.db.UpgradeToRedByUser(request.Context(), params.Data.UserID)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to upgraded user.", err)
		return
	}
	util.InfoLogger.Printf("Successfully upgraded user for id: %s", params.Data.UserID)

	writer.WriteHeader(http.StatusNoContent)
	util.InfoLogger.Printf("Upgrade finished.")
	return
}
