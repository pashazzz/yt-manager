# yt-manager

Сервис для просмотра YouTube-плейлистов как сериалов с отслеживанием прогресса.

## Стек

| Слой | Технология |
|------|-----------|
| Backend | Go 1.22 |
| БД | CloverDB v2 (встраиваемая NoSQL) |
| HTTP | Gin |
| YouTube | yt-dlp via go-ytdlp |
| Auth | Tailscale (заголовки `Tailscale-User-*`) |

## Структура

```
yt-manager/
├── cmd/server/
│   ├── main.go          # точка входа, wiring зависимостей
│   └── server.go        # graceful HTTP-сервер
├── internal/
│   ├── db/db.go         # инициализация CloverDB
│   ├── models/          # Profile, Show, Episode
│   ├── repository/      # CRUD-операции для каждой модели
│   ├── ytdlp/           # обёртка над yt-dlp binary
│   ├── handlers/        # HTTP-хендлеры (shows, episodes)
│   └── middleware/      # извлечение профиля из Tailscale-заголовков
└── go.mod
```

## Быстрый старт

```bash
# 1. Установить зависимости
go mod tidy

# 2. Убедиться, что yt-dlp доступен (или он будет скачан автоматически)
which yt-dlp   # macOS: brew install yt-dlp

# 3. Запустить сервер
go run ./cmd/server

# Сервер слушает :8080, данные хранятся в ./data/
```

## Переменные окружения

| Переменная | По умолчанию | Описание |
|-----------|-------------|---------|
| `ADDR` | `:8080` | Адрес и порт сервера |
| `DATA_DIR` | `./data` | Директория для файлов БД |

## API

### Шоу

```
POST   /api/v1/shows          — добавить плейлист (загружает эпизоды через yt-dlp)
GET    /api/v1/shows          — список шоу текущего профиля
GET    /api/v1/shows/:id      — шоу + список эпизодов
DELETE /api/v1/shows/:id      — удалить шоу и все эпизоды
```

### Эпизоды

```
POST   /api/v1/episodes/:id/progress — сохранить позицию просмотра
```

## Примеры запросов (curl)

```bash
# Добавить плейлист
curl -X POST http://localhost:8080/api/v1/shows \
  -H "Content-Type: application/json" \
  -H "Tailscale-User-Login: me@example.com" \
  -H "Tailscale-User-Name: Me" \
  -d '{"playlistUrl": "https://www.youtube.com/playlist?list=PLxxxxxx"}'

# Список шоу
curl http://localhost:8080/api/v1/shows \
  -H "Tailscale-User-Login: me@example.com"

# Эпизоды шоу
curl http://localhost:8080/api/v1/shows/<show-id> \
  -H "Tailscale-User-Login: me@example.com"

# Сохранить прогресс (на 95-й секунде, не досмотрено)
curl -X POST http://localhost:8080/api/v1/episodes/<episode-id>/progress \
  -H "Content-Type: application/json" \
  -H "Tailscale-User-Login: me@example.com" \
  -d '{"currentTime": 95.0, "isWatched": false}'
```

## Идентификация через Tailscale

При работе через `tailscale serve` / `tailscale funnel` Tailscale автоматически
добавляет заголовки `Tailscale-User-Login` и `Tailscale-User-Name`.  
В dev-режиме (вне Tailscale) middleware подставляет `dev@local` как заглушку.
