# Multi-stage build для оптимізації розміру образу
FROM golang:1.23-alpine AS builder

# Встановлюємо необхідні пакети для збірки
RUN apk add --no-cache git gcc musl-dev

# Встановлюємо робочу директорію
WORKDIR /app

# Копіюємо go.mod та go.sum для кешування залежностей
COPY go.mod go.sum ./
RUN go mod download

# Копіюємо вихідний код
COPY . .

# Збираємо бінарники
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o wg-orbit-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o wg-orbit-client ./cmd/client

# Фінальний образ
FROM alpine:latest

# Встановлюємо необхідні пакети
RUN apk add --no-cache \
    wireguard-tools \
    iptables \
    ca-certificates \
    tzdata

# Створюємо користувача для безпеки
RUN addgroup -g 1001 wg-orbit && \
    adduser -D -u 1001 -G wg-orbit wg-orbit

# Створюємо необхідні директорії
RUN mkdir -p /etc/wg-orbit /var/lib/wg-orbit /var/log/wg-orbit && \
    chown -R wg-orbit:wg-orbit /etc/wg-orbit /var/lib/wg-orbit /var/log/wg-orbit

# Копіюємо бінарники з builder stage
COPY --from=builder /app/wg-orbit-server /usr/local/bin/
COPY --from=builder /app/wg-orbit-client /usr/local/bin/

# Копіюємо конфігураційні файли
COPY configs/ /etc/wg-orbit/

# Встановлюємо права доступу
RUN chmod +x /usr/local/bin/wg-orbit-server /usr/local/bin/wg-orbit-client

# Відкриваємо порти
EXPOSE 8080 51820/udp

# Встановлюємо користувача
USER wg-orbit

# Встановлюємо робочу директорію
WORKDIR /home/wg-orbit

# Точка входу за замовчуванням
CMD ["wg-orbit-server", "run", "--config", "/etc/wg-orbit/server.yaml"]

# Мітки для метаданих
LABEL maintainer="WireGuard Orbit Team" \
      description="DevOps-friendly WireGuard management tool" \
      version="1.0.0"