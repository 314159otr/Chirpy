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
	platform := os.Getenv("PLATFORM")
	if dbURL == "" {
		log.Fatal("missing DB_URL")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening the database %w\n", err)
	}

	dbQueries := database.New(dbConn)
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
	}
	serveMux := http.NewServeMux()

	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	serveMux.HandleFunc("GET /api/healthz", handlerReadiness)
	serveMux.HandleFunc("POST /api/users", apiCfg.handlerUsers)
	serveMux.HandleFunc("POST /api/chirps", apiCfg.handlerChirps)

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


