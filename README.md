# OrderKeeper

**OrderKeeper** — это веб-сервис на Go для обработки, хранения и получения данных о заказах. Сервис демонстрирует чистую архитектуру, работу с базой данных, кеширование, взаимодействие с брокером сообщений (Apache Kafka) и предоставляет HTTP API для доступа к данным.

---

## Технологии

- **Язык**: Go
- **База данных**: PostgreSQL
- **Брокер сообщений**: Apache Kafka
- **HTTP Роутер**: `go-chi/chi`
- **ORM**: `gorm`
- **Клиент Kafka**: `segmentio/kafka-go`
- **Тестирование**: `stretchr/testify`, `uber-go/mock`

---

## Структура проекта

```
.
├── cmd/main.go              # Точка входа в приложение
├── internal/                # Внутренняя логика, не предназначенная для экспорта
│   ├── cache/cache.go       # Реализация in-memory кеша
│   ├── db/db.go             # Инициализация и подключение к БД
│   ├── handler/             # Обработчики HTTP-запросов (слой API)
│   ├── kafka/consumer.go    # Kafka-консьюмер
│   ├── models/              # Структуры данных (модели)
│   ├── repository/          # Слой доступа к данным (работа с БД)
│   └── service/             # Слой бизнес-логики
├── web/                     # Статические файлы для веб-интерфейса (HTML, JS)
├── docs/                    # Файлы для Swagger-документации
├── docker-compose.yml       # Файл для запуска всего проекта в Docker
├── Dockerfile               # Dockerfile для сборки приложения
├── go.mod                   # Зависимости проекта
└── README.md                # Этот файл
```

---

## Как запустить

### Предварительные требования

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Go](https://golang.org/) (для локальной разработки и тестов)

### Запуск с помощью Docker

Это самый простой и рекомендуемый способ.

1. **Клонируйте репозиторий:**

  ```bash
    git clone https://github.com/tiltdepressed/OrderKeeper.git
    cd ./OrderKeeper
  ```

2. **Создайте файл `.env`:**
    Скопируйте `.env.example` в новый файл с именем `.env`.

  ```bash
    cp .env.example .env
  ```

  Файл уже содержит все необходимые настройки для работы с `docker-compose`.

3. **Запустите проект:**

  ```bash
    docker-compose up --build
  ```

  Эта команда соберет образ вашего Go-приложения и запустит три контейнера: `app` (сам сервис), `db` (PostgreSQL) и `kafka` (брокер сообщений).

4. **Проект готов к работе!**
    - **Веб-интерфейс**: [http://localhost:8080/](http://localhost:8080/)
    - **Swagger API Docs**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

### Локальная разработка

1. **Запустите зависимости:**
    Вы можете запустить только базу данных и Kafka через Docker:

   ```bash
    docker-compose up -d db kafka
   ```

  Ключ `-d` запускает контейнеры в фоновом режиме.

2. **Создайте файл `.env`** как описано выше.

3. **Запустите приложение:**

   ```bash
    go run ./cmd/main.go
   ```

---

## Тестирование

Для запуска тестов и проверки покрытия кода:

1. **Сгенерируйте моки** (если вы вносили изменения в интерфейсы):

   ```bash
    go generate ./...
   ```

2. **Запустите тесты:**

   ```bash
    go test ./... -v -cover
   ```

---

## Использование API

### Получить заказ по ID

- **Endpoint**: `GET /order/{id}`
- **Описание**: Возвращает полные данные заказа по его уникальному идентификатору. Данные сначала ищутся в кеше, и если их там нет, то в базе данных.
- **Пример запроса**:

   ```bash
    curl http://localhost:8080/order/b563feb7b2b84b6test
   ```

- **Ответы**:
  - `200 OK`: Заказ найден, в теле ответа — JSON с данными заказа.
  - `404 Not Found`: Заказ с таким ID не найден.
  - `500 Internal Server Error`: Произошла внутренняя ошибка.

### Добавление заказов

Основной способ добавления заказов — отправка сообщения в топик Kafka `orders`. Сервис автоматически обработает сообщение и сохранит заказ.

- **Топик Kafka**: `orders`
- **Пример сообщения (JSON)**:

```json
 {
    "order_uid": "b563feb7b2b84b6test",
    "track_number": "WBILMTESTTRACK",
    "entry": "WBIL",
    "delivery": {
       "name": "Test Testov",
       "phone": "+9720000000",
       "zip": "2639809",
       "city": "Kiryat Mozkin",
       "address": "Ploshad Mira 15",
       "region": "Kraiot",
       "email": "test@gmail.com"
    },
    "payment": {
       "transaction": "b563feb7b2b84b6test",
       "request_id": "",
       "currency": "USD",
       "provider": "wbpay",
       "amount": 1817,
       "payment_dt": 1637907727,
       "bank": "alpha",
       "delivery_cost": 1500,
       "goods_total": 317,
       "custom_fee": 0
    },
    "items": [
       {
          "chrt_id": 9934930,
          "track_number": "WBILMTESTTRACK",
          "price": 453,
          "rid": "ab4219087a764ae0btest",
          "name": "Mascaras",
          "sale": 30,
          "size": "0",
          "total_price": 317,
          "nm_id": 2389212,
          "brand": "Vivienne Sabo",
          "status": 202
       }
    ],
    "locale": "en",
    "internal_signature": "",
    "customer_id": "test",
    "delivery_service": "meest",
    "shardkey": "9",
    "sm_id": 99,
    "date_created": "2021-11-26T06:22:19Z",
    "oof_shard": "1"
 }
 ```

- **Пример запроса**:

   ```bash
    echo '{"order_uid":"b563feb7b2b84b6test","track_number":"WBILMTESTTRACK","entry":"WBIL","delivery":{"name":"Test Testov","phone":"+9720000000","zip":"2639809","city":"Kiryat Mozkin","address":"Ploshad Mira 15","region":"Kraiot","email":"test@gmail.com"},"payment":{"transaction":"b563feb7b2b84b6test","request_id":"","currency":"USD","provider":"wbpay","amount":1817,"payment_dt":1637907727,"bank":"alpha","delivery_cost":1500,"goods_total":317,"custom_fee":0},"items":[{"chrt_id":9934930,"track_number":"WBILMTESTTRACK","price":453,"rid":"ab4219087a764ae0btest","name":"Mascaras","sale":30,"size":"0","total_price":317,"nm_id":2389212,"brand":"Vivienne Sabo","status":202}],"locale":"en","internal_signature":"","customer_id":"test","delivery_service":"meest","shardkey":"9","sm_id":99,"date_created":"2021-11-26T06:22:19Z","oof_shard":"1"}' | docker-compose exec -T kafka kafka-console-producer --broker-list kafka:29092 --topic orders
   ```

---

## Автор

- **tiltdepressed** - [GitHub Профиль](https://github.com/tiltdepressed)
