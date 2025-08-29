# 🚀 Повний приклад використання WireGuard Orbit

## 📋 Передумови

Переконайтеся, що у вас встановлені залежності:
```bash
# Ubuntu/Debian
sudo apt update && sudo apt install -y wireguard-tools iproute2 iptables

# CentOS/RHEL/Fedora
sudo dnf install -y wireguard-tools iproute iptables
```

## 🖥️ Налаштування сервера

### 1. Ініціалізація сервера
```bash
# Ініціалізуємо WireGuard інтерфейс wg0
sudo ./bin/wg-orbit-server init --interface wg0

# Або з кастомними параметрами
sudo ./bin/wg-orbit-server init --interface wg0 --port 51820 --network 10.0.0.0/24
```

### 2. Запуск сервера
```bash
# Запускаємо сервер (за замовчуванням на порту 8080)
sudo ./bin/wg-orbit-server run

# Або з кастомною конфігурацією
sudo ./bin/wg-orbit-server run --config configs/server.yaml
```

### 3. Додавання користувачів
```bash
# Додаємо нового користувача
./bin/wg-orbit-server user add dev1

# Генеруємо одноразовий токен для реєстрації
./bin/wg-orbit-server user token dev1
# Виведе щось на кшталт: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## 💻 Налаштування клієнта

### 1. Реєстрація клієнта
```bash
# Реєструємося на сервері з токеном
./bin/wg-orbit-client enroll --server http://your-server:8080 --token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Або з кастомним ім'ям клієнта
./bin/wg-orbit-client enroll --server http://your-server:8080 --token TOKEN --name laptop-dev
```

### 2. Підключення до VPN
```bash
# Піднімаємо WireGuard з'єднання
sudo ./bin/wg-orbit-client up

# Перевіряємо статус
./bin/wg-orbit-client status
```

### 3. Відключення
```bash
# Опускаємо WireGuard з'єднання
sudo ./bin/wg-orbit-client down
```

## 🔧 Приклад повного сценарію

### На сервері:
```bash
# 1. Ініціалізація
sudo ./bin/wg-orbit-server init --interface wg0 --port 51820 --network 10.0.0.0/24

# 2. Запуск сервера в фоні
sudo nohup ./bin/wg-orbit-server run > server.log 2>&1 &

# 3. Додавання користувачів
./bin/wg-orbit-server user add laptop
./bin/wg-orbit-server user add phone
./bin/wg-orbit-server user add tablet

# 4. Генерація токенів
LAPTOP_TOKEN=$(./bin/wg-orbit-server user token laptop)
PHONE_TOKEN=$(./bin/wg-orbit-server user token phone)
TABLET_TOKEN=$(./bin/wg-orbit-server user token tablet)

echo "Laptop token: $LAPTOP_TOKEN"
echo "Phone token: $PHONE_TOKEN"
echo "Tablet token: $TABLET_TOKEN"
```

### На клієнті (laptop):
```bash
# 1. Реєстрація
./bin/wg-orbit-client enroll --server http://10.0.1.100:8080 --token $LAPTOP_TOKEN --name laptop

# 2. Підключення
sudo ./bin/wg-orbit-client up

# 3. Перевірка статусу
./bin/wg-orbit-client status
# Виведе інформацію про з'єднання, IP адресу, статус

# 4. Тестування з'єднання
ping 10.0.0.1  # ping до сервера
```

## 🐳 Використання з Docker

### Запуск сервера:
```bash
# Збираємо образ
docker build -t wg-orbit .

# Запускаємо сервер
docker run -d --name wg-orbit-server \
  --privileged \
  --cap-add=NET_ADMIN \
  --cap-add=SYS_MODULE \
  -p 8080:8080 \
  -p 51820:51820/udp \
  -v /lib/modules:/lib/modules:ro \
  -v $(pwd)/data:/app/data \
  wg-orbit server run
```

### Або з docker-compose:
```bash
docker-compose up -d
```

## 📊 Моніторинг

```bash
# Перевірка статусу WireGuard інтерфейсу
sudo wg show

# Перегляд логів сервера
tail -f server.log

# REST API запити
curl http://localhost:8080/api/peers
curl http://localhost:8080/api/status
```

## 🔐 Безпека

- Токени мають обмежений час життя (за замовчуванням 24 години)
- Клієнти автоматично оновлюють токени кожні 12 годин
- Можливість відкликання доступу через видалення користувача
- JWT автентифікація для всіх API запитів

Тепер у вас є повнофункціональна WireGuard Orbit інфраструктура! 🎉