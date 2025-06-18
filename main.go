package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shale0315/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(db)

	const port = "8080"

	apiCfg := apiConfig{dbQueries: dbQueries, platform: platform}
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("./app")))))
	mux.HandleFunc("GET /api/healthz", readinessEndpoint)
	mux.Handle("POST /admin/reset", http.HandlerFunc(apiCfg.resetHandler))
	mux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.metricsHandler))
	mux.Handle("POST /api/chirps", http.HandlerFunc(apiCfg.ChirpHandler))
	mux.Handle("GET /api/chirps", http.HandlerFunc(apiCfg.SortChirpHandler))
	mux.Handle("POST /api/users", http.HandlerFunc(apiCfg.createUserHandler))
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
