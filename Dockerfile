# Используем официальный образ Go
FROM golang:1.23

# Устанавливаем gcc для CGO
RUN apt-get update && apt-get install -y gcc sqlite3 ca-certificates && update-ca-certificates

# Рабочая директория внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum и устанавливаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем всё остальное
COPY . .

#CGO — это мост между Go и C. go-sqlite3 использует нативную C-библиотеку SQLite → требует включённый CGO.
ENV CGO_ENABLED=1
# Сборка бинарника
RUN go build -o forum ./cmd && chmod +x forum

# Порт
EXPOSE 8080

# Запуск
CMD ["./forum"]
