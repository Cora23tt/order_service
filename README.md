# Order Service

**Order Service** — это REST API сервер, разработанный на языке Go. Сервис предоставляет функциональность управления заказами, пользователями и продуктами, с поддержкой JWT-аутентификации, разграничения прав доступа по ролям, экспортом данных и документированным API через Swagger.

## 📦 Возможности

- Аутентификация и авторизация по JWT
- CRUD-операции над заказами и продуктами
- Ограничение доступа на основе ролей (`user` / `admin`)
- Экспорт заказов в формате JSON и CSV с фильтрацией
- Swagger-документация всех эндпоинтов

## 🏗️ Используемые технологии

- Web-фреймворк: Gin
- База данных: PostgreSQL (`pgx`)
- Аутентификация: JWT
- Swagger-документация: swaggo
- Dependency Injection: `dig`
- Логирование: `zap`

## 📁 Структура проекта
```
.
├── cmd/server/                   # Точка входа: main.go
├── internal/
│   ├── rest/
│   │   ├── handlers/             # HTTP-обработчики (auth, user, order, product)
│   │   ├── middleware/           # JWT-валидация и другие middleware
│   │   └── rest.go               # Установка и маршрутов API
│   ├── repository/               # Реализация доступа к БД
│   └── usecase/                  # Бизнес-логика (сервисы)
├── pkg/
│   ├── db/                       # Инициализация подключения к PostgreSQL
│   ├── errors/                   # Централизованные ошибки
│   └── logger/                   # Логгер на базе zap
├── web/                          # Статические файлы (аватары, изображения и т.д.)
├── docs/                         # Swagger-генерация
├── .env                          # Переменные окружения
```

## 🔧 Установка и запуск

1. Клонировать репозиторий:

```bash
git clone https://github.com/your-username/order_service.git
cd order_service
```

2. Создать файл `.env` и указать переменные окружения:

```
DATABASE_URL=postgres://user:password@localhost:5432/orderdb
JWT_SECRET=your_jwt_secret_key
PORT=8080
LOG_LEVEL=debug
```

3. Запустить сервер:

```bash
go run cmd/server/main.go
```

Сервер будет доступен по адресу `http://localhost:8080`.

## 🔐 Авторизация

- Авторизация осуществляется через Bearer JWT токен.
- Для получения токена необходимо сперва зареистрироваться, после выполнить POST-запрос на `/api/v1/auth/signin`, передав номер телефона и пароль.
- Полученный токен указывается в заголовке каждого защищённого запроса:

```
Authorization: Bearer <token>
```

## 📘 Документация API

Документация доступна по адресу:  
`http://localhost:8080/swagger/index.html`

Документация содержит описание всех маршрутов, параметров, тел запроса и ответов, включая примеры. Используется аннотация `@swaggo` в коде для автогенерации.

## 📤 Примеры эндпоинтов

- `POST /api/v1/auth/signin` — вход и получение JWT
- `GET /api/v1/orders` — список заказов текущего пользователя
- `POST /api/v1/orders` — создание нового заказа
- `PUT /api/v1/orders/{id}` — обновление статуса заказа (admin)
- `GET /api/v1/orders/export` — экспорт заказов в JSON
- `GET /api/v1/orders/export/csv` — экспорт заказов в CSV

## 👤 Роли

- **user** — доступ к своим заказам
- **admin** — управление всеми заказами, пользователями, продуктами и экспортом данных

## 🧪 Тестирование

<sub>На момент защиты тесты не успели реализовать, но мы владеем навыками написания unit тестов и планируюем дополнить проект покрытием. 😇</sub>
