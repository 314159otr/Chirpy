package main

import(
	"net/http"
	"time"
	"database/sql"

	"github.com/314159otr/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, req * http.Request) {
	type resBody struct{
		Token string `json:"token"`
	}
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error getting token", err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(req.Context(), token)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusUnauthorized, "error getting user from refresh token", err)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error getting user from refresh token", err)
		return
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making JWT", err)
		return
	}

	respondWithJSON(w, http.StatusOK, resBody{
		Token: jwt,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, req * http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error getting token", err)
		return
	}

	if err := cfg.db.RevokeRefreshToken(req.Context(), token); err != nil {
		respondWithError(w, http.StatusInternalServerError, "error revoking refresh token", err)
	}

	w.WriteHeader(http.StatusNoContent)
}
