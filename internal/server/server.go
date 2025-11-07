package server

import (
	"net/http"

	"github.com/fathersson/wb-demo-service/internal/config"
)

// NewServer — возвращает http.Server
func NewServer(cfg config.HttpServer) *http.Server {
	mux := http.NewServeMux()

	// маршрутизация
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	return &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}
}
