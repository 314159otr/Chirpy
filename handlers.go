package main

import(
	"net/http"
	"fmt"
	"encoding/json"
	"strings"
	"time"
	"database/sql"

	"github.com/google/uuid"

	"github.com/314159otr/Chirpy/internal/auth"
	"github.com/314159otr/Chirpy/internal/database"
)

func cleanBody(body string, profaneWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		if _, found := profaneWords[strings.ToLower(word)]; found {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, req * http.Request) {
	type reqBody struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int `json:"expires_in_seconds"`
	}
	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)
	data := reqBody{}
	if err := decoder.Decode(&data); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	if data.ExpiresInSeconds == 0 || data.ExpiresInSeconds > 3600 {
		data.ExpiresInSeconds = 3600
	}

	user, err := cfg.db.GetUserByEmail(req.Context(), data.Email)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting user", err)
		return
	}
	matched, err := auth.CheckPasswordHash(data.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error checking password", err)
		return
	}
	if matched == false {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	jwt, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Duration(data.ExpiresInSeconds) * time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making JWT", err)
		return
	}
	respondWithJSON(w, http.StatusOK, User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			Token:     jwt,
	})
}

func (cfg *apiConfig) handlerUsersPost(w http.ResponseWriter, req * http.Request) {
	type reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)
	data := reqBody{}
	if err := decoder.Decode(&data); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(data.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password", err)
		return
	}
	createUserParams := database.CreateUserParams{
		Email:          data.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.db.CreateUser(req.Context(), createUserParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating user", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
	})
}

func handlerReadiness(w http.ResponseWriter, req * http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, req * http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	html := fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
	w.Write([]byte(html))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req * http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "you are not on DEV", nil)
		return
	}
	cfg.fileserverHits.Store(0)
	err := cfg.db.DeleteUsers(req.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting users", err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reseted to 0"))
}
