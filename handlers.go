package main

import(
	"net/http"
	"fmt"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
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


func (cfg *apiConfig) handlerUsers(w http.ResponseWriter, req * http.Request) {
	type reqBody struct {
		Email string `json:"email"`
	}
	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)
	data := reqBody{}
	if err := decoder.Decode(&data); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	user, err := cfg.db.CreateUser(req.Context(), data.Email)
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
