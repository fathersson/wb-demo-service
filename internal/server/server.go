package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/fathersson/wb-demo-service/internal/cache"
	"github.com/fathersson/wb-demo-service/internal/config"
	"github.com/fathersson/wb-demo-service/internal/repository"
)

// NewServer — возвращает http.Server
func NewServer(cfg config.HttpServer, cache cache.CacheInterface, db repository.OrderRepository) *http.Server {
	mux := http.NewServeMux()

	// Get order
	mux.HandleFunc("/order/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Получаем ID заказа из URL
		id := strings.TrimPrefix(r.URL.Path, "/order/")
		if id == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Получаем заказ из кэша
		order, ok := cache.GetCache(id)
		if ok {
			// Указываем, что ответ будет в формате JSON
			w.Header().Set("Content-Type", "application/json")
			// Код ответа 200 OK
			w.WriteHeader(http.StatusOK)
			// Кодируем структуру в JSON и отправляем клиенту
			json.NewEncoder(w).Encode(order)
			log.Printf("Заказ %s в кеше найден", id)
			log.Printf("Тело заказа: %+v", order)
			return
		}
		log.Printf("Заказ %s в кеше не нашли", id)

		// Получаем заказ из БД если в кеше нет
		order, err := db.GetOrderById(r.Context(), id)
		if err != nil {
			log.Printf("Заказ %s в БД не нашли", id)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			// http.Error(w, err.Error(), http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "not found",
			})
			return
		}
		log.Printf("Заказ %s в БД найден", id)
		log.Printf("Тело заказа: %+v", order)

		// Сохраняем заказ в кэш
		cache.SetCache(id, order)

		// Указываем, что ответ будет в формате JSON
		w.Header().Set("Content-Type", "application/json")
		// Код ответа 200 OK
		w.WriteHeader(http.StatusOK)
		// Кодируем структуру в JSON и отправляем клиенту
		json.NewEncoder(w).Encode(order)

	})

	// Раздача статических файлов
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	log.Println("Сервер будет запущен на", cfg.Port)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: CORS(mux),
	}
}

// Midleware обработка запроса для браузера
func CORS(mux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем все источники
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Возвращаем управление серверу
		mux.ServeHTTP(w, r)
	})
}
