package main

import(
	"net/http"
	"encoding/json"
	"time"
	"database/sql"
	"log"

	"github.com/314159otr/Chirpy/internal/database"
	"github.com/314159otr/Chirpy/internal/auth"

	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, req * http.Request) {
	type reqBody struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error getting JWT", err)
		return
	}

	log.Printf("token: %v\n", token)
	log.Printf("jwtSecret: %v\n", cfg.jwtSecret)
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error validating JWT", err)
		return
	}
	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)
	data := reqBody{}
	if err := decoder.Decode(&data); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	if len(data.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := cleanBody(data.Body, profaneWords)

	createChirpParams := database.CreateChirpParams{
		Body:   cleaned,
		UserID: userID,
	}
	chirp, err := cfg.db.CreateChirp(req.Context(), createChirpParams)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error creating chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	})
}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, req * http.Request) {
	chirps, err := cfg.db.GetChirps(req.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error getting chirps", err)
		return
	}
	var responseChirps []Chirp
	for _, chirp := range chirps {
		responseChirps = append(responseChirps, Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		})
	}
	respondWithJSON(w, http.StatusOK, responseChirps)
}

func (cfg *apiConfig) handlerChirpsGetByID(w http.ResponseWriter, req * http.Request) {
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid uuid", err)
		return
	}

	chirp, err := cfg.db.GetChirpByID(req.Context(), chirpID)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "error getting chirp", err)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error getting chirp", err)
		return
	}
	respondWithJSON(w, http.StatusOK, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	})
}
