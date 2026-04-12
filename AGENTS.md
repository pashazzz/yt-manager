# AGENTS.md — yt-manager

Этот файл — инструкция для AI-агентов (Copilot, Gemini, Cursor и т.д.),
работающих с этим репозиторием.

---

## Обзор проекта

**yt-manager** — Go-сервис для просмотра YouTube-плейлистов в формате сериалов.
Хранит прогресс просмотра, синхронизирует между устройствами через Tailscale.

| Параметр | Значение |
|---|---|
| Язык | Go 1.25 |
| HTTP | Gin v1.12 |
| БД | CloverDB v2 alpha (`github.com/ostafen/clover/v2 v2.0.0-alpha.3`) |
| YouTube | yt-dlp binary через `github.com/lrstanley/go-ytdlp v1.3.5` |
| Порт | `:8090` (ENV: `ADDR`) |
| Данные | `./data/` (ENV: `DATA_DIR`) |

---

## Структура проекта

```
yt-manager/
├── cmd/server/
│   ├── main.go          # точка входа: DI-wiring, роутер, graceful shutdown
│   └── server.go        # startServer() — gin в отдельной горутине
├── internal/
│   ├── db/
│   │   └── db.go        # Open() — инициализирует CloverDB и создаёт коллекции
│   ├── models/
│   │   ├── profile.go   # Profile {ID, Name}
│   │   ├── show.go      # Show {ID, Title, PlaylistURL, OwnerID, CreatedAt}
│   │   └── episode.go   # Episode {ID, ShowID, VideoID, Title, Duration, CurrentTime, IsWatched, OrderIndex}
│   ├── repository/
│   │   ├── profile_repo.go  # Upsert, FindByID
│   │   ├── show_repo.go     # Create, FindByOwner, FindByID, Delete + stringField helper
│   │   └── episode_repo.go  # BulkCreate, FindByShow, FindByID, UpdateProgress, DeleteByShow
│   ├── ytdlp/
│   │   └── ytdlp.go     # Client.FetchPlaylist() — вызывает yt-dlp binary
│   ├── handlers/
│   │   ├── shows.go     # ShowHandler: POST/GET/DELETE /shows, GET /shows/:id
│   │   └── episodes.go  # EpisodeHandler: POST /episodes/:id/progress
│   └── middleware/
│       └── profile.go   # Profile() — Tailscale-заголовки → gin.Context
├── AGENTS.md
├── Makefile
├── README.md
├── go.mod
└── go.sum
```

---

## Команды сборки и запуска

```bash
# Загрузить зависимости (всегда первый шаг после изменений go.mod)
go mod tidy

# Запустить сервер для разработки
go run ./cmd/server

# Скомпилировать бинарь в ./bin/yt-manager
make build

# Статический анализ
go vet ./...

# Полная проверка (tidy + vet)
make check
```

> **Тестов нет** (пока). Для проверки используй `curl`-команды из `Makefile`
> (`make curl-list`, `make curl-add LIST_ID=...`, и т.д.).

---

## API

Все маршруты имеют prefix `/api/v1`.

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/shows` | Добавить плейлист, загрузить эпизоды через yt-dlp |
| `GET` | `/shows` | Список шоу текущего профиля |
| `GET` | `/shows/:id` | Шоу + его эпизоды |
| `DELETE` | `/shows/:id` | Удалить шоу и все эпизоды |
| `POST` | `/episodes/:id/progress` | Сохранить позицию просмотра |

---

## Архитектурные паттерны и правила

### 1. Идентификация пользователя — только через middleware

Никогда не читай `Tailscale-User-Login` напрямую в хендлере.
Всегда используй `middleware.GetProfile(c)`:

```go
// ✅ правильно
profile := middleware.GetProfile(c)

// ❌ неправильно
login := c.GetHeader("Tailscale-User-Login")
```

### 2. CloverDB v2 alpha — особенности

**`FindFirst` может вернуть `ErrDocumentNotExist`**, а не `(nil, nil)`.
Всегда обрабатывай оба случая:

```go
doc, err := r.db.FindFirst(q)
if err != nil {
    if errors.Is(err, clover.ErrDocumentNotExist) {
        return nil, nil  // просто не нашли — это нормально
    }
    return nil, err      // реальная ошибка
}
if doc == nil {
    return nil, nil
}
```

**Кастомный `_id`** передаётся через ключ `"_id"` в map при создании документа:

```go
doc := document.NewDocumentOf(map[string]any{
    "_id":   myUUID,   // CloverDB использует это как ObjectId
    "field": value,
})
```

**Чтение полей** — CloverDB возвращает `interface{}`, всегда используй type assertion с проверкой:

```go
// ✅ правильно (с fallback)
name, _ := doc.Get("name").(string)

// ❌ неправильно (паника если nil)
name := doc.Get("name").(string)
```

**Числовые поля** — при чтении из БД могут быть `float64`, `int` или `float32`.
Используй готовые хелперы из `episode_repo.go`: `floatField`, `intField`, `boolField`.

### 3. Репозитории — один файл на модель

- `stringField` определён в `show_repo.go` и доступен всему пакету `repository`.
- `floatField`, `intField`, `boolField` определены в `episode_repo.go`.
- Не дублируй хелперы между файлами одного пакета.

### 4. Ownership check в хендлерах

Перед любой мутацией проверяй, что ресурс принадлежит текущему профилю:

```go
show, err := h.shows.FindByID(showID)
if show == nil || show.OwnerID != profile.ID {
    c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
    return
}
```

### 5. yt-dlp клиент — через context

`FetchPlaylist` принимает `context.Context`. Всегда передавай `c.Request.Context()` из хендлера, чтобы отмена HTTP-запроса прерывала вызов бинаря:

```go
info, err := h.ytClient.FetchPlaylist(c.Request.Context(), url)
```

### 6. Новые коллекции

При добавлении новой модели:
1. Добавь константу в `internal/db/db.go` (`CollectionXxx = "xxx"`).
2. Добавь коллекцию в цикл `Open()` там же.
3. Создай файл `internal/repository/xxx_repo.go`.
4. Добавь репо в `cmd/server/main.go`.

---

## Стиль кода

- **Форматирование**: `gofmt` / `goimports`. Не коммить неформатированный код.
- **Именование**: Go-конвенции (`camelCase` для методов, `PascalCase` для экспортируемого).
- **Ошибки**: всегда оборачивай через `fmt.Errorf("context: %w", err)`.
- **Логирование**: используй стандартный `log` пакет (не `fmt.Println`).
- **Комментарии к экспортируемым символам**: обязательны в формате `// FuncName ...`.
- **Нет глобальных переменных** — всё передаётся через dependency injection.

---

## Environment variables

| Переменная | По умолчанию | Описание |
|---|---|---|
| `ADDR` | `:8090` | Адрес и порт HTTP-сервера |
| `DATA_DIR` | `./data` | Директория для файлов CloverDB |
| `GIN_MODE` | `debug` | `release` для продакшна |

---

## Известные особенности / gotchas

### CloverDB v2.0.0-alpha.3 + Badger backend

- БД — это **директория**, а не один файл. `DATA_DIR` должна быть доступна для записи.
- Не запускай два экземпляра с одной `DATA_DIR` — Badger использует lock-файл.
- При обновлении до стабильного CloverDB v2 проверь: API `FindFirst`, `ErrDocumentNotExist`, и сигнатуру `Insert()` могут измениться.

### go-ytdlp v1.3.5

- `gytdlp.Install(ctx, nil)` возвращает `(*ytdlp.Ytdlp, error)` — игнорируй первое значение если не нужен объект.
- Сам yt-dlp вызывается через `exec.CommandContext` с флагами `--flat-playlist --dump-single-json`.
- Если плейлист приватный или yt-dlp не установлен — хендлер вернёт `502 Bad Gateway`.

### Tailscale headers в dev-режиме

- Вне Tailscale заголовки отсутствуют → middleware подставляет `dev@local`.
- В продакшне через `tailscale serve` заголовки добавляются автоматически.
- Не задавай фиктивные заголовки в продакшн-окружении.
