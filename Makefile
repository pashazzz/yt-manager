.PHONY: run build tidy vet check frontend build-all clean

# ─── Backend ──────────────────────────────────────────────────────────────────

# Запустить сервер (фронт раздаётся из internal/web/dist/)
run:
	go run ./cmd/server

# Скомпилировать бинарь в ./bin/yt-manager
build:
	go build -o bin/yt-manager ./cmd/server

# Загрузить и привести в порядок зависимости
tidy:
	go mod tidy

# Статический анализ
vet:
	go vet ./...

# Полная проверка: tidy + vet
check: tidy vet
	@echo "✓ backend checks passed"

# ─── Frontend ─────────────────────────────────────────────────────────────────

# Установить npm-зависимости и собрать фронтенд → internal/web/dist/
frontend:
	cd frontend && npm install && npm run build

# Запустить Vite dev-сервер (порт 5173, прокси на :8090)
frontend-dev:
	cd frontend && npm run dev

# ─── Combined ─────────────────────────────────────────────────────────────────

# Собрать всё: фронт + бинарь (production)
build-all: frontend build
	@echo "✓ build-all complete → bin/yt-manager"

# ─── Cleanup ──────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/ data/ internal/web/dist/*
	touch internal/web/dist/.gitkeep

# ─── API smoke tests (curl) ───────────────────────────────────────────────────

# make curl-add LIST_ID=PLxxxxxx
curl-add:
	curl -s -X POST http://localhost:8090/api/v1/shows \
		-H "Content-Type: application/json" \
		-H "Tailscale-User-Login: dev@local" \
		-H "Tailscale-User-Name: Dev User" \
		-d '{"playlistUrl":"https://www.youtube.com/playlist?list=$(LIST_ID)"}' | jq

curl-list:
	curl -s http://localhost:8090/api/v1/shows \
		-H "Tailscale-User-Login: dev@local" | jq

# make curl-show ID=<show-id>
curl-show:
	curl -s http://localhost:8090/api/v1/shows/$(ID) \
		-H "Tailscale-User-Login: dev@local" | jq

# make curl-progress ID=<episode-id> T=95
curl-progress:
	curl -s -X POST http://localhost:8090/api/v1/episodes/$(ID)/progress \
		-H "Content-Type: application/json" \
		-H "Tailscale-User-Login: dev@local" \
		-d '{"currentTime":$(T),"isWatched":false}' | jq
