package main

import(
	"net/http"
	"log"
	"sync/atomic"
	"os"
	"database/sql"

	"github.com/314159otr/Chirpy/internal/database"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	const port = "8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("missing DB_URL")
	}
	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("missing JWT_SECRET")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening the database %v\n", err)
	}

	dbQueries := database.New(dbConn)
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		jwtSecret:      jwtSecret,
	}
	serveMux := http.NewServeMux()

	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	serveMux.HandleFunc("GET /api/healthz", handlerReadiness)
	serveMux.HandleFunc("POST /api/users", apiCfg.handlerUsersPost)
	serveMux.HandleFunc("PUT /api/users", apiCfg.handlerUsersPut)
	serveMux.HandleFunc("POST /api/login", apiCfg.handlerLogin)

	serveMux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	serveMux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)

	serveMux.HandleFunc("POST /api/chirps", apiCfg.handlerChirps)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsGet)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpsGetByID)
	serveMux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerChirpsDeleteByID)

	serveMux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerPolkaWebhooks)

	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)


	server := &http.Server{
		Addr: ":" + port,
		Handler: serveMux,
	}
	log.Printf("Starting the server on port: %s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}


