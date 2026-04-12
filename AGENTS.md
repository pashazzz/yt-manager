# AGENTS.md — yt-manager

Этот файл — инструкция для AI-агентов (Copilot, Gemini, Cursor и т.д.),
работающих с этим репозиторием.

---

## Обзор проекта

**yt-manager** — Go-сервис для просмотра YouTube-плейлистов в формате сериалов.
Хранит прогресс просмотра, синхронизирует между устройствами через Tailscale.

| Параметр | Значение |
|---|---|
| Язык (backend) | Go 1.25 |
| HTTP | Gin v1.12 |
| БД | CloverDB v2 alpha (`github.com/ostafen/clover/v2 v2.0.0-alpha.3`) |
| YouTube | yt-dlp binary через `github.com/lrstanley/go-ytdlp v1.3.5` |
| Frontend | React 18 + Vite 5 + TypeScript |
| Порт | `:8090` (ENV: `ADDR`) |
| Данные | `./data/` (ENV: `DATA_DIR`) |

---

## Структура проекта

```
yt-manager/
├── cmd/server/
│   ├── main.go          # точка входа: DI-wiring, роутер, SPA-сервер, graceful shutdown
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
│   ├── middleware/
│   │   └── profile.go   # Profile() — Tailscale-заголовки → gin.Context
│   └── web/
│       ├── embed.go     # //go:embed all:dist — встраивает фронтенд в бинарь
│       └── dist/        # сюда Vite собирает фронтенд (gitignored, кроме .gitkeep)
├── frontend/
│   ├── src/
│   │   ├── api/client.ts        # fetch-обёртки над всеми API эндпоинтами
│   │   ├── types/index.ts       # TypeScript-типы (Show, Episode, ShowDetail, …)
│   │   ├── components/
│   │   │   ├── VideoPlayer.tsx  # YouTube IFrame + seekTo + heartbeat + мобильный overlay
│   │   │   ├── EpisodeList.tsx  # Сайдбар со списком эпизодов
│   │   │   ├── ShowCard.tsx     # Карточка шоу с прогрессом
│   │   │   └── AddShowModal.tsx # Модалка добавления плейлиста
│   │   ├── pages/
│   │   │   ├── ShowsPage.tsx    # Главная: сетка шоу
│   │   │   └── ShowPage.tsx     # Страница просмотра: плеер + сайдбар
│   │   ├── App.tsx              # React Router: / и /shows/:id
│   │   ├── main.tsx             # точка входа React
│   │   └── index.css            # design system (CSS custom properties)
│   ├── public/
│   │   └── icon.svg             # favicon
│   ├── vite.config.ts           # outDir: ../internal/web/dist
│   └── package.json
├── AGENTS.md
├── Makefile
├── README.md
├── go.mod
└── go.sum
```

---

## Команды

```bash
# Разработка (два терминала)
go run ./cmd/server                    # бэкенд на :8090
cd frontend && npm run dev             # фронтенд на :5173 с прокси → :8090

# Production-сборка (один бинарь)
make build-all                         # фронт → internal/web/dist/ + go build

# Отдельно
make frontend                          # только npm build
make build                             # только go build

# Проверки
make check                             # go mod tidy + go vet
go vet ./...

# Чистка
make clean                             # удаляет bin/, data/, сбрасывает dist/
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

## Как работает раздача фронтенда

```
npm run build  →  internal/web/dist/   (//go:embed all:dist в internal/web/embed.go)
                        ↓
go build       →  bin/yt-manager  ←  фронт встроен в бинарь
```

- Gin раздаёт статику через `http.FileServer(http.FS(webFS))`
- Все пути, NOT начинающиеся с `/api/`, → `index.html` (SPA fallback)
- В dev-режиме Vite-сервер (`:5173`) проксирует `/api` → `:8090`
- `internal/web/dist/.gitkeep` обязательно должен оставаться в репозитории — иначе `go build` упадёт из-за пустой `//go:embed all:dist`

### Добавление новой API-группы

1. Зарегистрируй маршруты в `cmd/server/main.go` с префиксом `/api/v1/...`
2. Фронтенд после этого автоматически не будет перехватывать этот путь (SPA-фолбек применяется только если нет реального файла)

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

**Кастомный `_id`** должен быть **валидным UUID** (формат `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`).
CloverDB v2 alpha **отвергает любую другую строку** (email, slug и т.п.) с ошибкой `invalid _id: <value>`.

```go
// ✅ правильно — UUID генерирует uuid.NewString()
doc := document.NewDocumentOf(map[string]any{
    "_id":   uuid.NewString(), // или другой валидный UUID
    "field": value,
})

// ❌ неправильно — email не является UUID, будет ошибка
doc := document.NewDocumentOf(map[string]any{
    "_id": "user@example.com", // invalid _id!
})
```

**Для моделей, где натуральный ключ не является UUID** (например, Tailscale-логин — это email),
храни натуральный ключ в отдельном поле и делай lookup по нему:

```go
// Профиль: email хранится в поле "login", _id генерирует CloverDB
doc := document.NewDocumentOf(map[string]any{
    // "_id" не задаём — CloverDB сгенерирует валидный UUID автоматически
    "login": profile.ID, // email из Tailscale-заголовка
    "name":  profile.Name,
})
// Поиск по login, а не по _id:
q := query.NewQuery(col).Where(query.Field("login").Eq(email))
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

### 6. Новые коллекции (backend)

При добавлении новой модели:
1. Добавь константу в `internal/db/db.go` (`CollectionXxx = "xxx"`).
2. Добавь коллекцию в цикл `Open()` там же.
3. Создай файл `internal/repository/xxx_repo.go`.
4. Добавь репо в `cmd/server/main.go`.

### 7. Новые страницы (frontend)

1. Добавь новый компонент в `frontend/src/pages/`.
2. Зарегистрируй роут в `frontend/src/App.tsx`.
3. Новые API-методы добавляй в `frontend/src/api/client.ts`.
4. CSS-классы — в `frontend/src/index.css` (дизайн-система на CSS custom properties).

---

## Стиль кода

### Go
- **Форматирование**: `gofmt` / `goimports`. Не коммить неформатированный код.
- **Именование**: Go-конвенции (`camelCase` для методов, `PascalCase` для экспортируемого).
- **Ошибки**: всегда оборачивай через `fmt.Errorf("context: %w", err)`.
- **Логирование**: используй стандартный `log` пакет (не `fmt.Println`).
- **Комментарии к экспортируемым символам**: обязательны в формате `// FuncName ...`.
- **Нет глобальных переменных** — всё передаётся через dependency injection.

### TypeScript / React
- **Компоненты**: один файл = один компонент, PascalCase имена файлов.
- **Стили**: только через CSS-классы из `index.css`. Никакого inline-style кроме динамических значений (`width: \`${progress}%\``).
- **API-вызовы**: только через `src/api/client.ts`, не `fetch` напрямую в компонентах.
- **Типы**: все типы данных в `src/types/index.ts`.

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

### //go:embed и пустой dist/

- `internal/web/dist/.gitkeep` **обязателен** в репозитории.
- Без него `go build` падает: `//go:embed all:dist` не может встроить несуществующую директорию.
- После `make clean` `.gitkeep` восстанавливается автоматически.

### Tailscale headers в dev-режиме

- Вне Tailscale заголовки отсутствуют → middleware подставляет `dev@local`.
- В продакшне через `tailscale serve` заголовки добавляются автоматически.
- Не задавай фиктивные заголовки в продакшн-окружении.

### VideoPlayer — мобильные устройства

- `playsinline: 1` в `playerVars` обязателен, иначе iOS открывает видео в системном плеере.
- На мобильных показывается overlay «Начать просмотр» — первый тап выступает user gesture для автоплея.
- Heartbeat (10 сек) стартует только после первого `onPlay` события.
- При размонтировании компонента (`useEffect` cleanup) используется `navigator.sendBeacon` для сохранения последней позиции без блокировки навигации.
