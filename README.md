# Event Managment Backend

Backend-приложение для управления мероприятиями, построенное на микросервсиной архитектуре.

Цель проекта — изучение построения распределённых систем на Go, взаимодействия сервисов через gRPC и организации единой точки входа с помощью GraphQL Gateway.

## Архитектура 

Проект разделён на несколько независимых сервисов.

### Auth service 

Отвечает за:

- регистрацию пользователей;
- аутенфикацию;
- выдачу JWT-токенов;
- назначение модерации и установки отдел, к которму относится модератор.

### Event Service

Отвечает за:

- создание мероприятий;
- получение информации о мероприятиях;
- управление данными мероприятий.

### Attendee Service

Отвечает за:

- регистрацию пользователей на мероприятия;
- хранение информации об участниках.

### GraphQL Gateway

Является единой точкой входа для клиента.

Получает GraphQL-запросы и взаимодействует с внутренними сервисами через gRPC.

## Используемые технологии

- Go
- GraphQL
- gqlgen
- gRPC
- Protocol Buffers
- PostgreSQL
- GORM
- JWT
- Docker
- Docker Compose
- MinIO


## Архитектура взаимодействия

```text
                    Client
                       │
                GraphQL Gateway
                       │
      ┌────────────────┼────────────────┐
      │                │                │
      ▼                ▼                ▼
 Auth Service    Event Service   Attendee Service
      │                │                │
      └────────────────┴────────────────┘
                       │
                PostgreSQL / MinIO
```

## Структура проекта

```text
.
├── attendee/
├── auth/
├── docker
├── event/
├── graphql/
├── pkg
├── docker-compose.yml
└── README.md
```

## Запуск

### Клонирование репозитория

```bash
git clone https://github.com/svladislav00-qq/event-microservices
cd event-microservices
```

### Запуск

```bash
docker compose up --build -d
```

### Назначение администратора

```bash
docker exec event_postgres psql -U postgres -d event_ms_accounts -c "UPDATE accounts SET role = 'admin', updated_at = NOW() WHERE email = 'user@example.com';"
```
Использовать после регистрации пользователя для создания пользователя с ролью 'admin' для назначения модерации.

## Дальнейшее развитие

Планируется:

- добавить юнит тесты для всех сервисов;
- добавить интеграционные тесты;
- реализовать централизованное логирование;
