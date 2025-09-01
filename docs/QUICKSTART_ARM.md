# Швидкий старт wg-orbit на ARM пристроях

## 🚀 Швидке розгортання (5 хвилин)

### Raspberry Pi / Orange Pi з Docker

```bash
# 1. Встановлення Docker
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker $USER
sudo reboot

# 2. Запуск wg-orbit
docker run -d \
  --name wg-orbit \
  --cap-add NET_ADMIN \
  --restart unless-stopped \
  -p 8080:8080 \
  -p 51820:51820/udp \
  -v wg-orbit-data:/etc/wg-orbit \
  ghcr.io/artem/wg-orbit:latest

# 3. Перевірка статусу
docker logs wg-orbit
curl http://localhost:8080/api/v1/health
```

### Додавання користувача

```bash
# Додати нового peer
docker exec wg-orbit wg-orbit-server user add mydevice

# Отримати токен для підключення
docker exec wg-orbit wg-orbit-server user token mydevice
```

### Підключення клієнта

```bash
# На клієнтському пристрої
wget https://github.com/artem/wg-orbit/releases/latest/download/wg-orbit-linux-arm64.tar.gz
tar -xzf wg-orbit-linux-arm64.tar.gz
sudo mv wg-orbit-client /usr/local/bin/

# Реєстрація з токеном
wg-orbit-client enroll --server YOUR_SERVER_IP:8080 --token YOUR_TOKEN

# Підключення
sudo wg-orbit-client up
```

## 📋 Підтримувані пристрої

| Пристрій | Статус | Docker образ |
|----------|--------|-------------|
| Raspberry Pi 4 | ✅ Рекомендовано | `linux/arm64` |
| Raspberry Pi 3 | ✅ Підтримується | `linux/arm/v7` |
| Orange Pi 5 | ✅ Рекомендовано | `linux/arm64` |
| Orange Pi PC | ✅ Підтримується | `linux/arm/v7` |
| NVIDIA Jetson | ✅ Підтримується | `linux/arm64` |

## 🔧 Налаштування системи

```bash
# Увімкнення IP forwarding
echo 'net.ipv4.ip_forward=1' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Встановлення WireGuard tools (якщо потрібно)
sudo apt update
sudo apt install -y wireguard-tools
```

## 🐛 Швидке усунення проблем

### Проблема: Container не запускається
```bash
# Перевірка логів
docker logs wg-orbit

# Перевірка прав
ls -la /var/run/docker.sock
sudo chmod 666 /var/run/docker.sock
```

### Проблема: Не працює мережа
```bash
# Перевірка модуля WireGuard
sudo modprobe wireguard
lsmod | grep wireguard

# Перевірка iptables
sudo iptables -L -n
```

### Проблема: Повільна робота
```bash
# Перевірка ресурсів
free -h
df -h
htop

# Оптимізація для SD карти
sudo systemctl disable swap
```

## 📚 Детальна документація

Для повної інформації дивіться [ARM_DEPLOYMENT.md](ARM_DEPLOYMENT.md)

## 🆘 Підтримка

- GitHub Issues: https://github.com/artem/wg-orbit/issues
- Документація: https://github.com/artem/wg-orbit/docs
- Telegram: @wg-orbit-support