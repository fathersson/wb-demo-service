# wb-demo-service
Демонстрационный сервис с Kafka, PostgreSQL, кешем (WbTex L0) 

* Чтобы у вас работали переменные окружения нужно выполнить команду cp .env.example .env

Описание в разработке:

что делает сервис
как его запустить
какие зависимости
пример запроса GET /order/<id>
пример JSON заказа

PostgreSQL доступен на localhost:5432
Kafka UI на http://localhost:8080
Kafka на localhost:9092
Zookeeper на localhost:2181

# Подключение к PostgreSQL через pgAdmin

Добавь сервер:
Host: postgres (имя сервиса в docker-compose)
Port: 5432
Database: wb_orders
Username: wb_user
Password: wb_pass

# Рекомендуемая структура репозитория

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
docker ps                # проверяем, что какие контейнеры запущены
docker-compose down -v   # останавливаем и удаляем старые контейнеры и тома

# kafka topik create

docker exec -it wb_kafka kafka-topics \
  --create \
  --topic orders \
  --bootstrap-server kafka:9092 \
  --partitions 3 \
  --replication-factor 1

# sql tables create

  CREATE TABLE orders (
    order_uid VARCHAR(50) PRIMARY KEY,
    track_number VARCHAR(50),
    entry VARCHAR(20),
    locale VARCHAR(10),
    internal_signature TEXT,
    customer_id VARCHAR(50),
    delivery_service VARCHAR(50),
    shardkey VARCHAR(10),
    sm_id INT,
    date_created TIMESTAMP,
    oof_shard VARCHAR(10)
);

CREATE TABLE delivery (
    order_uid VARCHAR(50) PRIMARY KEY REFERENCES orders(order_uid),
    name VARCHAR(255),
    phone VARCHAR(50),
    zip VARCHAR(20),
    city VARCHAR(100),
    address TEXT,
    region VARCHAR(100),
    email VARCHAR(100)
);

CREATE TABLE payment (
    transaction VARCHAR(50) PRIMARY KEY REFERENCES orders(order_uid),
    request_id VARCHAR(50),
    currency VARCHAR(10),
    provider VARCHAR(50),
    amount NUMERIC(12,2),
    payment_dt TIMESTAMP,
    bank VARCHAR(50),
    delivery_cost NUMERIC(12,2),
    goods_total NUMERIC(12,2),
    custom_fee NUMERIC(12,2)
);

CREATE TABLE items (
    chrt_id BIGINT PRIMARY KEY,
    order_uid VARCHAR(50) REFERENCES orders(order_uid),
    track_number VARCHAR(50),
    price NUMERIC(12,2),
    rid VARCHAR(50),
    name VARCHAR(255),
    sale NUMERIC(5,2),
    size VARCHAR(10),
    total_price NUMERIC(12,2),
    nm_id BIGINT,
    brand VARCHAR(100),
    status INT
);

# SQL values

INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
VALUES ('b563feb7b2b84b6test','WBILMTESTTRACK','WBIL','en','','test','meest','9',99,'2021-11-26T06:22:19Z','1');

INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
VALUES ('b563feb7b2b84b6test','Test Testov','+9720000000','2639809','Kiryat Mozkin','Ploshad Mira 15','Kraiot','test@gmail.com');

INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
VALUES ('b563feb7b2b84b6test','','USD','wbpay','1817',to_timestamp(1637907727),'alpha','1500','317','0');

INSERT INTO items (chrt_id, order_uid, price, rid, name, sale, size, total_price, nm_id, brand, status)
VALUES
('9934930','b563feb7b2b84b6test','453','ab4219087a764ae0btest','Mascaras','30','0','317','2389212','Vivienne Sabo','202'),
('9934931','b563feb7b2b84b6test','500','cd1234567a764ae0btest','Lipstick','10','M','450','2389213','BrandX','201');
