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
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(db)

	const port = "8080"

	apiCfg := apiConfig{dbQueries: dbQueries}
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("./app")))))
	mux.HandleFunc("GET /api/healthz", readinessEndpoint)
	mux.Handle("POST /admin/reset", http.HandlerFunc(apiCfg.resetHandler))
	mux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.metricsHandler))
	mux.Handle("POST /api/validate_chirp", http.HandlerFunc(handlerChripsValidate))
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
