package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseWriter обертывает http.ResponseWriter для перехвата статус-кода и размера ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Logger возвращает middleware для логирования запросов и ответов
func Logger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Оборачиваем ResponseWriter
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Вызываем следующий обработчик
			next.ServeHTTP(rw, r)

			// Вычисляем время выполнения
			duration := time.Since(start)

			// Логируем на уровне Info
			logger.Info("HTTP Request",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Duration("duration", duration),
				zap.Int("status_code", rw.statusCode),
				zap.Int("response_size", rw.size),
			)
		}

		return http.HandlerFunc(fn)
	}
}
