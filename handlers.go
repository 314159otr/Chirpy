package main

import(
	"net/http"
	"log"
	"fmt"
	"encoding/json"
	"strings"
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

func handlerValidateChirp(w http.ResponseWriter, req * http.Request) {
	type reqBody struct {
		Body string `json:"body"`
	}
	type responsePayload struct {
		CleanedBody string `json:"cleaned_body"`
	}

	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)
	data := reqBody{}
	err := decoder.Decode(&data)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	if len(data.Body) > 140 {
		log.Printf("Chirp is too long", data.Body)
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := cleanBody(data.Body, profaneWords)
	respondWithJSON(w, http.StatusOK, responsePayload{
		CleanedBody: cleaned,
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
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reseted to 0"))
}
