# yt-manager

Сервис для просмотра YouTube-плейлистов как сериалов с отслеживанием прогресса, мульти-теггингом и синхронизацией через Tailscale.

## Стек

| Слой | Технология |
|------|-----------|
| Backend | Go 1.25, Gin v1.12 |
| БД | CloverDB v2 alpha |
| YouTube | yt-dlp via go-ytdlp |
| Frontend | React 18 + Vite + TypeScript |
| Auth | Tailscale (`Tailscale-User-*` заголовки) |

## Структура

```
yt-manager/
├── cmd/server/
│   ├── main.go          # точка входа, DI, SPA-сервер
│   └── server.go        # HTTP-сервер с graceful shutdown
├── internal/
│   ├── db/db.go         # инициализация CloverDB
│   ├── models/          # Profile, Show, Episode
│   ├── repository/      # CRUD для каждой модели
│   ├── ytdlp/           # обёртка над yt-dlp binary
│   ├── handlers/        # HTTP-хендлеры (shows, episodes)
│   ├── middleware/       # Tailscale → Profile
│   └── web/
│       ├── embed.go     # //go:embed all:dist — встраивает фронтенд
│       └── dist/        # сюда Vite кладёт билд (gitignored, кроме .gitkeep)
└── frontend/            # React-приложение
    ├── src/
    │   ├── api/         # fetch-обёртки
    │   ├── components/  # ShowCard, EpisodeList, VideoPlayer, AddShowModal
    │   ├── pages/       # ShowsPage, ShowPage
    │   └── types/       # TypeScript-типы
    └── vite.config.ts   # outDir → ../internal/web/dist
```

## Быстрый старт

### Разработка (два терминала)

```bash
# Терминал 1 — бэкенд
go run ./cmd/server

# Терминал 2 — фронтенд с hot-reload (прокси на :8090)
cd frontend && npm install && npm run dev
# → открой http://localhost:5173
```

### Production (один бинарь)

```bash
# 1. Собрать фронтенд и бэкенд
make build-all

# 2. Запустить
./bin/yt-manager
# → http://localhost:8090
```

## Команды

| Команда | Описание |
|---------|---------|
| `make run` | Запустить бэкенд (с уже собранным фронтом) |
| `make frontend` | npm install + vite build → internal/web/dist/ |
| `make frontend-dev` | Vite dev-сервер на :5173 |
| `make build-all` | Фронт + Go бинарь в ./bin/yt-manager |
| `make check` | go mod tidy + go vet |
| `make clean` | Удалить bin/, data/, сбросить dist/ |

## Переменные окружения

| Переменная | По умолчанию | Описание |
|-----------|-------------|---------|
| `ADDR` | `:8090` | Адрес и порт сервера |
| `DATA_DIR` | `./data` | Директория для файлов БД |
| `GIN_MODE` | `debug` | `release` для продакшна |

## API

### Теги

```
GET    /api/v1/tags          — список тегов профиля (тег Default создаётся автоматически)
POST   /api/v1/tags          — создать новый тег
POST   /api/v1/tags/:id/delete — удалить (тег снимается с элементов, элементы не удаляются)
GET    /api/v1/tags/:id/items  — шоу и одиночные видео внутри тега
POST   /api/v1/tags/reorder — изменить порядок сортировки тегов (drag-n-drop)
POST   /api/v1/tags/:id/episodes — добавить одиночное видео напрямую в тег
```

### Шоу

```
POST   /api/v1/shows          — добавить плейлист или кастомное шоу (пустое)
GET    /api/v1/shows          — список всех шоу (legacy)
GET    /api/v1/shows/:id      — шоу + список эпизодов
POST   /api/v1/shows/:id/delete — удалить шоу и все эпизоды
POST   /api/v1/shows/:id/tags   — управление тегами шоу (мульти-теггинг)
POST   /api/v1/shows/:id/reverse — переключить сортировку эпизодов (старые/новые)
POST   /api/v1/shows/:id/episodes — добавить одиночное видео (для кастомных шоу)
POST   /api/v1/shows/:id/episodes/reorder — изменить сортировку эпизодов (drag-n-drop кастомных шоу)
```

### Эпизоды

```
POST   /api/v1/episodes/:id/progress — сохранить позицию просмотра
POST   /api/v1/episodes/:id/tags     — управление тегами одиночного видео
POST   /api/v1/episodes/:id/delete   — удалить видео (эпизод)
```

## Как раздаётся фронтенд

```
npm run build  ─→  internal/web/dist/   (//go:embed all:dist)
                          ↓
go build       ─→  bin/yt-manager  ←  фронт встроен в бинарь
```

Gin раздаёт статику через `http.FileServer(http.FS(webFS))`.  
Все маршруты, не начинающиеся с `/api/`, получают `index.html` (SPA fallback).  
В dev-режиме Vite проксирует `/api` → `:8090` напрямую.

## Идентификация через Tailscale

При работе через `tailscale serve` Tailscale добавляет заголовки автоматически.  
Для локальной разработки middleware подставляет `dev@local` как заглушку.

## Примеры curl

```bash
# Добавить плейлист
make curl-add LIST_ID=PLxxxxxx

# Список шоу
make curl-list

# Эпизоды шоу
make curl-show ID=<show-id>

# Сохранить прогресс
make curl-progress ID=<episode-id> T=95
```
