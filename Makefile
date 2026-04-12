.PHONY: run build tidy lint clean

# Запустить сервер в режиме разработки
run:
	go run ./cmd/server

# Скомпилировать бинарь
build:
	go build -o bin/yt-manager ./cmd/server

# Загрузить и проверить все зависимости
tidy:
	go mod tidy

# Статический анализ
vet:
	go vet ./...

# Удалить скомпилированные артефакты
clean:
	rm -rf bin/ data/

# Быстрая проверка компиляции без запуска
check: tidy vet
	@echo "✓ all checks passed"

# --- Для тестирования API вручную ---
# Добавить плейлист (заменить LIST_ID на реальный ID)
curl-add:
	curl -s -X POST http://localhost:8080/api/v1/shows \
		-H "Content-Type: application/json" \
		-H "Tailscale-User-Login: dev@local" \
		-H "Tailscale-User-Name: Dev User" \
		-d '{"playlistUrl":"https://www.youtube.com/playlist?list=$(LIST_ID)"}' | jq

curl-list:
	curl -s http://localhost:8080/api/v1/shows \
		-H "Tailscale-User-Login: dev@local" | jq

curl-show:
	curl -s http://localhost:8080/api/v1/shows/$(ID) \
		-H "Tailscale-User-Login: dev@local" | jq

curl-progress:
	curl -s -X POST http://localhost:8080/api/v1/episodes/$(ID)/progress \
		-H "Content-Type: application/json" \
		-H "Tailscale-User-Login: dev@local" \
		-d '{"currentTime":$(T),"isWatched":false}' | jq
