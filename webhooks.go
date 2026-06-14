package main

import(
	"net/http"
	"encoding/json"
	"database/sql"

	"github.com/google/uuid"

	"github.com/314159otr/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerPolkaWebhooks(w http.ResponseWriter, req * http.Request) {
	apiKey, err := auth.GetAPIKey(req.Header)
	if err != nil || apiKey != cfg.polkaApiKey {
		respondWithError(w, http.StatusUnauthorized, "Error getting APIKey", err)
		return
	}

	type reqBody struct{
		Event string `json:"event"`
		Data  struct{
			UserID string `json:"user_id"`
		} `json:"data"`
	}
	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)
	data := reqBody{}
	if err := decoder.Decode(&data); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}
	if data.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if data.Event == "user.upgraded" {
		userID, err := uuid.Parse(data.Data.UserID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid uuid", err)
			return
		}
		err = cfg.db.UpgradeUserIsChirpyRedByID(req.Context(), userID)
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "error user not found", err)
			return
		}
		if err != nil{
			respondWithError(w, http.StatusInternalServerError, "error upgrading user", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
