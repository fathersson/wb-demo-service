# wb-demo-service
Демонстрационный сервис с Kafka, PostgreSQL, кешем (WbTex L0) 

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
