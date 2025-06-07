package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	const port = "8080"

	apiCfg := apiConfig{}
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("./app")))))
	mux.HandleFunc("GET /api/healthz", readinessEndpoint)
	mux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.metricsHandler))
	mux.Handle("POST /admin/reset", http.HandlerFunc(apiCfg.resetHandler))
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
