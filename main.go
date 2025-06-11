package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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
	mux.Handle("POST /admin/reset", http.HandlerFunc(apiCfg.resetHandler))
	mux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.metricsHandler))
	mux.Handle("POST /api/validate_chirp", http.HandlerFunc(handlerChripsValidate))
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
