# Используем официальный Go-образ для сборки
FROM golang:1.25.3-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем бинарник
RUN go build -o football-alice cmd/alice/main.go

# Минимальный образ для запуска
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS
RUN apk --no-cache add ca-certificates

# Копируем бинарник из стадии сборки
COPY --from=builder /app/football-alice /football-alice

# Открываем порт 8080
EXPOSE 8080

# Запуск приложения
ENTRYPOINT ["/football-alice"]
