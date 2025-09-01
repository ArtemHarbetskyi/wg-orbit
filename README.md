# WireGuard Orbit 🚀

**DevOps-friendly інструмент для автоматизації керування WireGuard**

WireGuard Orbit робить WireGuard простим для адміністрування: без ручного редагування конфігів, з автоматичною видачею IP, централізованим менеджментом peer'ів та ротацією ключів.

## 🎯 Особливості

- **Один бінарник** — `wg-orbit` з двома режимами: `server` та `client`
- **Автоматична видача IP** — IPAM light для автоматичного призначення адрес
- **REST API** — повне керування через HTTP API
- **JWT автентифікація** — безпечні токени з можливістю відкликання
- **Підтримка баз даних** — SQLite (за замовчуванням) або PostgreSQL (продакшн)
- **Моніторинг** — відстеження peer'ів, handshake, online/offline статусу
- **Docker-ready** — готові образи та Docker Compose конфігурації
- **Multi-arch підтримка** — ARM64/ARMv7 для Raspberry Pi, Orange Pi та інших мікрокомп'ютерів

## 🏗️ Архітектура

```
┌─────────────────┐    REST API     ┌─────────────────┐
│   wg-orbit      │◄───────────────►│   wg-orbit      │
│   (server)      │                 │   (client)      │
│                 │                 │                 │
│ ┌─────────────┐ │                 │ ┌─────────────┐ │
│ │ WireGuard   │ │                 │ │ WireGuard   │ │
│ │ Interface   │ │                 │ │ Interface   │ │
│ └─────────────┘ │                 │ └─────────────┘ │
│ ┌─────────────┐ │                 └─────────────────┘
│ │ SQLite/     │ │
│ │ PostgreSQL  │ │
│ └─────────────┘ │
└─────────────────┘
```

## 🚀 Швидкий старт

### Встановлення

```bash
# Клонування репозиторію
git clone https://github.com/your-org/wg-orbit.git
cd wg-orbit

# Збірка
make build

# Або використання Docker
make docker
```

## 📦 Встановлення залежностей

**Для використання wg-orbit потрібно попередньо встановити WireGuard.**

### Обов'язкові компоненти:

1. **WireGuard kernel module та tools:**
   - `wireguard-tools` (містить команди `wg`, `wg-quick`)
   - WireGuard kernel module (зазвичай вже включений в сучасні ядра Linux)

2. **Системні утиліти:**
   - `ip` (iproute2) - для керування мережевими інтерфейсами
   - `iptables` - для налаштування firewall правил

3. **Права доступу:**
   - `sudo` права для виконання команд `wg-quick`, `wg`, `ip`
   - Можливість створювати мережеві інтерфейси (CAP_NET_ADMIN)

### Встановлення за операційними системами:

#### RPi:

```bash
wget https://go.dev/dl/go1.22.6.linux-arm64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.6.linux-arm64.tar.gz
```
```bash
export PATH=$PATH:/usr/local/go/bin
```
```bash
source ~/.bashrc
```

#### Ubuntu/Debian:
```bash
sudo apt update
sudo apt install wireguard-tools iproute2 iptables
```

#### CentOS/RHEL/Fedora:
```bash
sudo dnf install wireguard-tools iproute iptables
# або для старіших версій:
sudo yum install wireguard-tools iproute iptables
```

#### Alpine Linux:
```bash
apk add wireguard-tools iptables
```

#### macOS:
```bash
brew install wireguard-tools
```

### Docker режим:

Якщо ви використовуєте Docker, всі залежності вже включені в образ:
- Dockerfile автоматично встановлює `wireguard-tools` та `iptables`
- Контейнер потребує `privileged: true` та `CAP_NET_ADMIN` для роботи з мережею

### Запуск сервера

```bash
# Ініціалізація сервера
./bin/wg-orbit-server init --interface wg0

# Запуск сервера
./bin/wg-orbit-server run
```

### Додавання клієнта

```bash
# На сервері: створення користувача
./bin/wg-orbit-server user add dev1

# Генерація токена
./bin/wg-orbit-server user token dev1

# На клієнті: реєстрація
./bin/wg-orbit-client enroll --server https://your-server:8080 --token <TOKEN>

# Підключення
./bin/wg-orbit-client up
```

## 📁 Структура проекту

```
wg-orbit/
├── cmd/                    # Точки входу
│   ├── server/             # Сервер
│   └── client/             # Клієнт
├── internal/               # Внутрішня логіка
│   ├── wg/                # WireGuard операції
│   ├── auth/              # JWT автентифікація
│   └── storage/           # Зберігання даних
├── api/                   # REST API
│   └── rest/              # HTTP handlers
├── configs/               # Конфігураційні файли
├── Dockerfile             # Docker образ
├── docker-compose.yml     # Локальна розробка
└── Makefile              # Автоматизація збірки
```

## 🛠️ Розробка

### Вимоги

- Go 1.21+
- Docker & Docker Compose
- WireGuard tools
- PostgreSQL (опціонально)

### Локальна розробка

```bash
# Налаштування середовища
make dev-setup

# Запуск бази даних
docker-compose up -d postgres

# Запуск сервера
make dev-server

# В іншому терміналі: тестування
curl http://localhost:8080/health
```

### Тестування

```bash
# Unit тести
make test

# Тести з покриттям
make test-cover

# Лінтинг
make lint

# Інтеграційні тести
make docker-run
# Тести в контейнерах...
make docker-stop
```

## 🔧 Підтримка ARM пристроїв

wg-orbit повністю підтримує ARM архітектури для розгортання на мікрокомп'ютерах:

### Підтримувані пристрої
- **Raspberry Pi 4/3** (ARM64/ARMv7)
- **Orange Pi 5/PC** (ARM64/ARMv7) 
- **NVIDIA Jetson Nano** (ARM64)
- **Rock Pi 4** (ARM64)

### Швидкий старт на ARM

```bash
# Docker (рекомендовано)
docker run -d \
  --name wg-orbit \
  --cap-add NET_ADMIN \
  --restart unless-stopped \
  -p 8080:8080 \
  -p 51820:51820/udp \
  -v wg-orbit-data:/etc/wg-orbit \
  ghcr.io/artem/wg-orbit:latest

# Нативна збірка
wget https://github.com/artem/wg-orbit/releases/latest/download/wg-orbit-linux-arm64.tar.gz
tar -xzf wg-orbit-linux-arm64.tar.gz
sudo mv wg-orbit-* /usr/local/bin/
```

### Документація для ARM
- 📖 [Повний гід по розгортанню на ARM](docs/ARM_DEPLOYMENT.md)
- 🚀 [Швидкий старт для ARM пристроїв](docs/QUICKSTART_ARM.md)
- 🐛 [Усунення проблем на ARM](docs/ARM_DEPLOYMENT.md#усунення-проблем)

## 📋 API Документація

### Основні ендпойнти

| Метод | Шлях | Опис |
|-------|------|------|
| `GET` | `/health` | Перевірка здоров'я |
| `POST` | `/api/v1/enroll` | Реєстрація клієнта |
| `GET` | `/api/v1/config` | Отримання конфігу |
| `POST` | `/api/v1/refresh` | Оновлення токена |
| `GET` | `/api/v1/peers` | Список peer'ів |
| `POST` | `/api/v1/peers` | Створення peer'а |
| `GET` | `/api/v1/peers/{id}` | Інформація про peer'а |
| `PUT` | `/api/v1/peers/{id}` | Оновлення peer'а |
| `DELETE` | `/api/v1/peers/{id}` | Видалення peer'а |

### Приклад використання

```bash
# Реєстрація клієнта
curl -X POST http://localhost:8080/api/v1/enroll \
  -H "Content-Type: application/json" \
  -d '{"token": "your-enrollment-token", "name": "my-device"}'

# Отримання конфігурації
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/config

# Список peer'ів
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/peers
```

## 🐳 Docker

### Збірка образу

```bash
make docker
```

### Запуск з Docker Compose

```bash
# Повне середовище (сервер + база + моніторинг)
make docker-run

# Перегляд логів
make docker-logs

# Зупинка
make docker-stop
```

### Сервіси в Docker Compose

- **wg-orbit-server** — основний сервер (порт 8080)
- **postgres** — база даних (порт 5432)
- **adminer** — веб-інтерфейс для БД (порт 8081)
- **prometheus** — збір метрик (порт 9090)
- **grafana** — візуалізація (порт 3000, admin/admin)

## ⚙️ Конфігурація

### Сервер (`configs/server.yaml`)

```yaml
server:
  host: "0.0.0.0"
  port: 8080

wireguard:
  interface: "wg0"
  listen_port: 51820
  network: "10.0.0.0/24"

storage:
  type: "sqlite"  # або "postgres"
  database: "/var/lib/wg-orbit/wg-orbit.db"

auth:
  jwt_secret: "your-secret-key"
  token_duration: "24h"
```

### Клієнт (`configs/client.yaml`)

```yaml
client:
  name: "my-device"
  interface: "wg0"

server:
  url: "https://your-server:8080"
  
auth:
  token_file: "/var/lib/wg-orbit/token"
  refresh_interval: "12h"
```

## 🔒 Безпека

- **JWT токени** з обмеженим терміном дії
- **TLS/HTTPS** для API (рекомендовано в продакшні)
- **Ротація ключів** WireGuard
- **Відкликання доступу** через API
- **Аудит логи** всіх операцій

## 📊 Моніторинг

### Метрики Prometheus

- `wg_orbit_peers_total` — загальна кількість peer'ів
- `wg_orbit_peers_online` — активні peer'і
- `wg_orbit_handshakes_total` — кількість handshake'ів
- `wg_orbit_api_requests_total` — API запити

### Логування

```bash
# Логи сервера
tail -f /var/log/wg-orbit/server.log

# Логи клієнта
tail -f /var/log/wg-orbit/client.log
```

## 🚀 Продакшн

### Systemd сервіс

```ini
[Unit]
Description=WireGuard Orbit Server
After=network.target

[Service]
Type=simple
User=wg-orbit
ExecStart=/usr/local/bin/wg-orbit-server run --config /etc/wg-orbit/server.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Nginx проксі

```nginx
server {
    listen 443 ssl http2;
    server_name wg-orbit.example.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 🤝 Внесок у проект

1. Fork репозиторію
2. Створіть feature branch (`git checkout -b feature/amazing-feature`)
3. Commit зміни (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Відкрийте Pull Request

## 📄 Ліцензія

MIT License - дивіться [LICENSE](LICENSE) файл для деталей.

## 🆘 Підтримка

- 📖 [Документація](https://github.com/your-org/wg-orbit/wiki)
- 🐛 [Issues](https://github.com/your-org/wg-orbit/issues)
- 💬 [Discussions](https://github.com/your-org/wg-orbit/discussions)

---

**WireGuard Orbit** — зробіть WireGuard простим! 🌟
