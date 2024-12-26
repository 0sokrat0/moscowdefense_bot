# Используем официальный образ Golang как базовый образ
FROM golang:1.23.4-alpine AS build

# Устанавливаем зависимости
RUN apk update && apk add --no-cache git sqlite

# Устанавливаем текущий рабочий каталог внутри контейнера
WORKDIR /app

# Копируем манифесты модулей Go
COPY go.mod go.sum ./

# Загружаем и кешируем модули Go
RUN go mod download

# Копируем файлы проекта в контейнер
COPY . .

# Собираем Go приложение
RUN go build -o /app/main ./cmd/bot/main.go

# Делаем исполняемый файл main выполняемым
RUN chmod +x /app/main

# Добавляем команду для проверки содержимого /app
RUN ls -l /app

# Используем официальный образ Python как базовый для веб-интерфейса
FROM python:3.8-slim AS web

# Устанавливаем зависимости
RUN pip install sqlite-web

# Копируем собранное Go приложение из предыдущего этапа
COPY --from=build /app /app

# Устанавливаем текущий рабочий каталог внутри контейнера
WORKDIR /app

# Добавляем команду для проверки содержимого /app
RUN ls -l /app

# Запускаем Go приложение и веб-интерфейс
CMD ["sh", "-c", "sqlite_web /app/database.db --host 0.0.0.0 --port 8080 & /app/main"]
