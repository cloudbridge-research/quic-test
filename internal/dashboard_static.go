package internal

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// StaticFileSystem предоставляет доступ к встроенным статическим файлам
type StaticFileSystem struct {
	httpFS http.FileSystem
}

// NewStaticFileSystem создает новую файловую систему для статических файлов
func NewStaticFileSystem() *StaticFileSystem {
	return &StaticFileSystem{
		httpFS: http.Dir("static"),
	}
}

// Open открывает файл
func (sfs *StaticFileSystem) Open(name string) (http.File, error) {
	return sfs.httpFS.Open(name)
}

// ServeStatic обрабатывает запросы к статическим файлам
func ServeStatic(w http.ResponseWriter, r *http.Request) {
	// Убираем префикс /static/ если он есть
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	if path == "" || path == "/" {
		path = "index.html"
	}
	
	// Проверяем расширение файла для правильного MIME типа
	ext := filepath.Ext(path)
	var contentType string
	
	switch ext {
	case ".html":
		contentType = "text/html; charset=utf-8"
	case ".css":
		contentType = "text/css; charset=utf-8"
	case ".js":
		contentType = "application/javascript; charset=utf-8"
	case ".json":
		contentType = "application/json"
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".svg":
		contentType = "image/svg+xml"
	case ".ico":
		contentType = "image/x-icon"
	case ".woff":
		contentType = "font/woff"
	case ".woff2":
		contentType = "font/woff2"
	case ".ttf":
		contentType = "font/ttf"
	case ".eot":
		contentType = "application/vnd.ms-fontobject"
	default:
		contentType = "text/plain; charset=utf-8"
	}
	
	w.Header().Set("Content-Type", contentType)
	
	// Читаем файл напрямую из файловой системы
	filePath := "static/" + path
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Если файл не найден, возвращаем index.html для SPA
		if path != "index.html" {
			content, err = os.ReadFile("static/index.html")
			if err != nil {
				http.Error(w, "File not found", http.StatusNotFound)
				return
			}
			// Устанавливаем правильный Content-Type для HTML
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
	}
	
	// Устанавливаем заголовки и отправляем содержимое
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
	w.Write(content)
}

// GetStaticFileList возвращает список доступных статических файлов
func GetStaticFileList() ([]string, error) {
	// Простая реализация - возвращаем основные файлы
	return []string{"index.html", "css/style.css", "js/app.js"}, nil
}

// StaticFileHandler создает обработчик для статических файлов
func StaticFileHandler() http.Handler {
	return http.HandlerFunc(ServeStatic)
}
