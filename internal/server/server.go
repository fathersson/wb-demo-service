package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/fathersson/wb-demo-service/internal/cache"
	"github.com/fathersson/wb-demo-service/internal/config"
)

// NewServer — возвращает http.Server
func NewServer(cfg config.HttpServer, cache *cache.Cache) *http.Server {
	mux := http.NewServeMux()

	// маршрутизация
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	// Get order
	mux.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Указываем, что ответ будет в формате JSON
		w.Header().Set("Content-Type", "application/json")

		// Код ответа 200 OK
		w.WriteHeader(http.StatusOK)

		// Получаем ID заказа из URL
		id := strings.TrimPrefix(r.URL.Path, "/order/")
		if id == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		data, ok := cache.GetCache(id) // Получаем заказ из кэша
		if !ok {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}

		// Кодируем структуру в JSON и отправляем клиенту
		json.NewEncoder(w).Encode(data)
	})

	return &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}
}

// func getOrderHandler(w http.ResponseWriter, r *http.Request, cache *cache.Cache) {
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Указываем, что ответ будет в формате JSON
// 	w.Header().Set("Content-Type", "application/json")

// 	// Код ответа 200 OK
// 	w.WriteHeader(http.StatusOK)

// 	// Получаем ID заказа из URL
// 	id := strings.TrimPrefix(r.URL.Path, "/order/")
// 	if id == "" {
// 		http.Error(w, "bad request", http.StatusBadRequest)
// 		return
// 	}

// 	data, ok := cache.GetCache(id) // Получаем заказ из кэша
// 	if !ok {
// 		http.Error(w, "order not found", http.StatusNotFound)
// 		return
// 	}

// 	// Кодируем структуру в JSON и отправляем клиенту
// 	json.NewEncoder(w).Encode(data)

// }
