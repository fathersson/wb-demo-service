package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/fathersson/wb-demo-service/internal/cache"
	"github.com/fathersson/wb-demo-service/internal/config"
	"github.com/fathersson/wb-demo-service/internal/repository"
)

// NewServer — возвращает http.Server
func NewServer(cfg config.HttpServer, cache *cache.Cache, db *sql.DB) *http.Server {
	mux := http.NewServeMux()

	// // маршрутизация
	// mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("Hello World!"))
	// })

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
			log.Printf("Заказ %s в кеше нашли", id)
			return
		}
		log.Printf("Заказ %s в кеше не нашли", id)

		// Получаем заказ из БД если в кеше нет
		order, err := repository.GetOrderById(r.Context(), db, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		log.Printf("Заказ %s в БД нашли", id)

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

	return &http.Server{
		Addr:    ":" + cfg.Port,
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
