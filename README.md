# wb-demo-service
Демонстрационный сервис с Kafka, PostgreSQL, кешем (WbTex L0) 

* Чтобы у вас работали переменные окружения нужно выполнить команду cp .env.example .env

запуск docker compose up -d

PostgreSQL доступен на localhost:5432
pgAdmin доступен на http://localhost:8081
Kafka UI на http://localhost:8080
Kafka на localhost:9092
Zookeeper на localhost:2181

4️⃣ Подключение к PostgreSQL через pgAdmin
В браузере открой http://localhost:8081
Вход: admin@example.com / admin (из .env)
Добавь сервер:
Host: postgres (имя сервиса в docker-compose)
Port: 5432
Database: wb_orders
Username: wb_user
Password: wb_pass






Описание в разработке:

что делает сервис
как его запустить
какие зависимости
пример запроса GET /order/<id>
пример JSON заказа

Рекомендуемая структура репозитория

wb-demo-service/
├── cmd/
│   └── app/
│       └── main.go           # вход в программу
│
├── internal/
│   ├── db/                   # логика работы с PostgreSQL
│   ├── kafka/                # consumer
│   ├── cache/                # in-memory cache
│   ├── http/                 # хендлеры и API
│   ├── models/               # структуры (Order, Delivery, Payment, Item)
│   └── config/               # чтение переменных окружения (.env)
│
├── web/                      # HTML/JS фронт
│   └── index.html
│
├── docker-compose.yml        # окружение: kafka + postgres + ui
├── .env.example              # пример конфигурации (без секретов)
├── go.mod
├── go.sum
└── README.md


# docker and docker-compose

docker-compose up -d     # запускаем контейнеры
docker-compose down      # осатнавливаем контейнеры
docker ps                # проверяем, что pgAdmin запущен
docker-compose down -v   # останавливаем и удаляем старые контейнеры и тома
