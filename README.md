# wb-demo-service

Демонстрационный сервис с Kafka, PostgreSQL, кешем в памяти и HTTP API.
Сервис предназначен для обработки заказов, поступающих из Kafka, сохранения их в базу данных и последующей выдачи через API с использованием кеширования.

## Возможности сервиса

* Читает JSON-заказы из Kafka топика `orders`.
* Сохраняет заказ в БД (PostgreSQL) и Кэш (map).
* После перезапуска сервиса подгружает кэш из бд.
* Возвращает заказ через `GET /order/<id>`.
* Повторный запрос обслуживается быстрее благодаря кешу.
* Поддерживает простейший HTML-интерфейс (папка `/web`).

---

## Запуск проекта

### 1. Подготовка переменных окружения

```
cp .env.example .env
```

### 2. Запуск всего окружения

```
docker compose up -d (запускаем контейнеры в фоновом режиме)
```

Остановка:

```
docker compose stop  # Просто останавливает контейнеры
docker compose down (останавливаем и удаляем контейнеры)
```

Удаление контейнеров и томов:

```
docker-compose down -v (тут удаляются еще и сохраненные данные)
```

Посмотреть активные контейнеры:

```
docker ps
```

Удалите конкретный именованный том:

```
docker volume ls - ищем название нужного тома

docker volume rm + название нужного тома
```

---

## Пример запроса к API

```
GET http://localhost:8082/order/<id>
```

Пример корректного ID:

```
GET http://localhost:8082/order/b563feb7b2b84b6test
```

---

## Доступные сервисы и порты

* PostgreSQL: `localhost:5432`
* Kafka: `localhost:9092`
* Kafka UI: `http://localhost:8080`
* Zookeeper: `localhost:2181`
* HTML интерфейс: `http://localhost:8082`

---

## Подключение к PostgreSQL через pgAdmin

Добавьте сервер:

```
Host: postgres
Port: 5432
Database: wb_orders
Username: wb_user
Password: wb_pass
```

---

## Структура репозитория

```
wb-demo-service/
├── cmd/
│   └── app/
│       └── main.go          # вход в приложение
├── internal/
│   ├── db/                  # подключение к PostgreSQL
│   ├── kafka/               # consumer Kafka
│   ├── cache/               # in-memory кеш
│   ├── server/              # HTTP-сервер и маршруты
│   ├── models/              # структуры данных
│   ├── config/              # конфигурация
│   └── repository/          # хранение логики чтения/записи данных
├── web/
│   └── index.html           # простой веб-интерфейс
├── docker-compose.yml
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

---

## Kafka: создание топика

```
docker exec -it wb_kafka kafka-topics \
  --create \
  --topic orders \
  --bootstrap-server kafka:9092 \
  --partitions 3 \
  --replication-factor 1
```

---

## SQL: таблицы

### orders

```
-- ================================
-- 1) Таблица orders
-- ================================
CREATE TABLE IF NOT EXISTS orders (
    order_uid          TEXT PRIMARY KEY,
    track_number       TEXT NOT NULL,
    entry              TEXT,
    locale             TEXT,
    internal_signature TEXT,
    customer_id        TEXT,
    delivery_service   TEXT,
    shardkey           TEXT,
    sm_id              INTEGER,
    date_created       TIMESTAMPTZ,
    oof_shard          TEXT
);

-- ================================
-- 2) Таблица delivery
-- ================================
CREATE TABLE IF NOT EXISTS delivery (
    order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name      TEXT NOT NULL,
    phone     TEXT NOT NULL,
    zip       TEXT NOT NULL,
    city      TEXT NOT NULL,
    address   TEXT NOT NULL,
    region    TEXT,
    email     TEXT
);

-- ================================
-- 3) Таблица payment
-- ================================
CREATE TABLE IF NOT EXISTS payment (
    order_uid     TEXT PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction   TEXT NOT NULL,
    request_id    TEXT,
    currency      TEXT NOT NULL,
    provider      TEXT NOT NULL,
    amount        INTEGER NOT NULL,
    payment_dt    BIGINT NOT NULL,
    bank          TEXT,
    delivery_cost INTEGER NOT NULL,
    goods_total   INTEGER NOT NULL,
    custom_fee    INTEGER
);

-- ================================
-- 4) Таблица items
-- ================================
CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid   TEXT NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id     INTEGER NOT NULL,
    track_number TEXT,
    price       INTEGER NOT NULL,
    rid         TEXT,
    name        TEXT NOT NULL,
    sale        INTEGER,
    size        TEXT,
    total_price INTEGER,
    nm_id       INTEGER,
    brand       TEXT,
    status      INTEGER
);

```

---

## Тестовые данные

```
INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature,
customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
VALUES ('b563feb7b2b84b6test','WBILMTESTTRACK','WBIL','en','','test','meest','9',99,'2021-11-26T06:22:19Z','1');

INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
VALUES ('b563feb7b2b84b6test','Test Testov','+9720000000','2639809','Kiryat Mozkin','Ploshad Mira 15','Kraiot','test@gmail.com');

INSERT INTO payment (transaction, request_id, currency, provider, amount,
payment_dt, bank, delivery_cost, goods_total, custom_fee)
VALUES ('b563feb7b2b84b6test','','USD','wbpay','1817',1637907727,'alpha','1500','317','0');

INSERT INTO items (chrt_id, order_uid, track_number, price, rid, name, sale, size,
total_price, nm_id, brand, status)
VALUES (9934930,'b563feb7b2b84b6test','WBILMTESTTRACK','453','ab4219087a764ae0btest','Mascaras',30,'0','317','2389212','Vivienne Sabo',202);
```

---

## Минимальный скрипт отправки сообщения в Kafka

```
go run producer.go (не реализован)
```

Где `producer.go` содержит JSON заказа и отправляет его в топик `orders`.

---

# Работа над ошибками

```
1) Избавился от неиспользуемых зависимостей - прогнал go mod tidy

2) В файл .env.example добавлены все переменные окружения, используемые в проекте, также поправлен config и docker-compose

3) Исправлены имена переменных, которые назывались именами импортируемых пакетов

4) Исправлены необработанные ошибки в коде

5) Реализовал graceful shutdown

6) Добавил инвалидацию кэша, теперь лимит данных 1000 ключей мапы

7-8) Реализована валидация данных всех важных полей в заказе с помощью пакета go-playground/validator и также закоментирован второй вариант - ручная проверка, с помощью сущности validate.go 

9) Реализован полноценный генератор данных, который поочередно посылает в топик валидное и битое сообщение

10)
```

---


