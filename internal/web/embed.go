package web

import (
	"embed"
	"io/fs"
)

// static содержит весь билд фронтенда, вложенный во время компиляции.
// Директория internal/web/dist/ создаётся командой: cd frontend && npm run build
//
//go:embed all:dist
var static embed.FS

// FS возвращает содержимое папки dist/ как fs.FS для использования
// в http.FileServer и gin.
func FS() fs.FS {
	sub, err := fs.Sub(static, "dist")
	if err != nil {
		panic("web: embed sub failed: " + err.Error())
	}
	return sub
}
