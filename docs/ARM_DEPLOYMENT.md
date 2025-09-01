# Розгортання wg-orbit на ARM пристроях

Цей документ містить рекомендації та інструкції для розгортання wg-orbit на мікрокомп'ютерах з ARM архітектурою.

## Підтримувані пристрої

### ✅ Повністю підтримувані
- **Raspberry Pi 4** (ARM64) - рекомендовано
- **Raspberry Pi 3** (ARMv7) - підтримується
- **Orange Pi 5** (ARM64) - рекомендовано
- **Orange Pi PC** (ARMv7) - підтримується
- **NVIDIA Jetson Nano** (ARM64)
- **Rock Pi 4** (ARM64)

### ⚠️ Обмежена підтримка
- **Raspberry Pi 2** (ARMv7) - мінімальні вимоги
- **Raspberry Pi Zero 2 W** (ARM64) - обмежена продуктивність

### ❌ Не підтримується
- **Raspberry Pi 1** (ARMv6) - застаріла архітектура
- **Raspberry Pi Zero** (ARMv6) - застаріла архітектура

## Системні вимоги

### Мінімальні вимоги
- **RAM**: 512 MB (рекомендовано 1 GB+)
- **Диск**: 2 GB вільного місця
- **ОС**: Linux з підтримкою WireGuard
- **Ядро**: Linux 5.6+ (з вбудованим WireGuard) або модуль wireguard-dkms

### Рекомендовані вимоги
- **RAM**: 2 GB+
- **Диск**: SSD або швидка SD карта (Class 10, U3)
- **Мережа**: Ethernet підключення (для стабільності)

## Встановлення

### Метод 1: Docker (рекомендовано)

```bash
# Встановлення Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Перезавантаження для застосування змін групи
sudo reboot

# Запуск wg-orbit
docker run -d \
  --name wg-orbit \
  --cap-add NET_ADMIN \
  --cap-add SYS_MODULE \
  -p 8080:8080 \
  -p 51820:51820/udp \
  -v /etc/wg-orbit:/etc/wg-orbit \
  -v /lib/modules:/lib/modules:ro \
  ghcr.io/artem/wg-orbit:latest
```

### Метод 2: Нативна збірка

```bash
# Встановлення залежностей
sudo apt update
sudo apt install -y wireguard-tools iptables iproute2 sqlite3

# Завантаження бінарників
wget https://github.com/artem/wg-orbit/releases/latest/download/wg-orbit-linux-arm64.tar.gz
# або для ARMv7:
# wget https://github.com/artem/wg-orbit/releases/latest/download/wg-orbit-linux-armv7.tar.gz

# Розпакування
tar -xzf wg-orbit-linux-*.tar.gz
sudo mv wg-orbit-server wg-orbit-client /usr/local/bin/
sudo chmod +x /usr/local/bin/wg-orbit-*

# Створення конфігураційних директорій
sudo mkdir -p /etc/wg-orbit
sudo chown $USER:$USER /etc/wg-orbit
```

## Оптимізація для ARM пристроїв

### 1. Налаштування SD карти (для Raspberry Pi)

```bash
# Збільшення розміру swap (якщо RAM < 2GB)
sudo dphys-swapfile swapoff
sudo sed -i 's/CONF_SWAPSIZE=100/CONF_SWAPSIZE=1024/' /etc/dphys-swapfile
sudo dphys-swapfile setup
sudo dphys-swapfile swapon

# Оптимізація для SD карти
echo 'tmpfs /tmp tmpfs defaults,noatime,nosuid,size=100m 0 0' | sudo tee -a /etc/fstab
echo 'tmpfs /var/tmp tmpfs defaults,noatime,nosuid,size=30m 0 0' | sudo tee -a /etc/fstab
```

### 2. Налаштування мережі

```bash
# Увімкнення IP forwarding
echo 'net.ipv4.ip_forward=1' | sudo tee -a /etc/sysctl.conf
echo 'net.ipv6.conf.all.forwarding=1' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Налаштування iptables для збереження правил
sudo apt install -y iptables-persistent
```

### 3. Оптимізація SQLite

Створіть файл `/etc/wg-orbit/server.yaml`:

```yaml
server:
  interface: wg0
  listen_port: 51820
  address: "10.0.0.1/24"
  
database:
  type: sqlite
  connection_string: "/etc/wg-orbit/wg-orbit.db"
  # Оптимізація для ARM пристроїв
  sqlite_options:
    journal_mode: WAL
    synchronous: NORMAL
    cache_size: -2000  # 2MB cache
    temp_store: MEMORY
    mmap_size: 67108864  # 64MB
```

## Моніторинг продуктивності

### Перевірка ресурсів

```bash
# Використання CPU та RAM
htop

# Використання диска
df -h
iostat -x 1

# Мережевий трафік
iftop -i wg0

# Статус WireGuard
sudo wg show
```

### Логи wg-orbit

```bash
# Docker логи
docker logs wg-orbit -f

# Системні логи (для нативної установки)
journalctl -u wg-orbit -f
```

## Усунення проблем

### Проблема: "exec format error"
**Причина**: Неправильна архітектура бінарника
**Рішення**: Переконайтеся, що використовуєте правильний бінарник:
- ARM64: `wg-orbit-linux-arm64`
- ARMv7: `wg-orbit-linux-armv7`

```bash
# Перевірка архітектури системи
uname -m
# aarch64 = ARM64
# armv7l = ARMv7
```

### Проблема: Повільна робота SQLite
**Рішення**: Використовуйте SSD або оптимізуйте налаштування:

```bash
# Перенесення БД на RAM диск (тимчасово)
sudo mkdir -p /tmp/wg-orbit
sudo cp /etc/wg-orbit/wg-orbit.db /tmp/wg-orbit/
# Оновіть connection_string в конфігурації
```

### Проблема: Високе навантаження на CPU
**Рішення**: Обмежте кількість одночасних підключень:

```yaml
server:
  max_peers: 50  # Зменшіть для слабших пристроїв
  keepalive_interval: 60  # Збільшіть інтервал
```

### Проблема: Проблеми з iptables на новіших версіях Raspberry Pi OS
**Причина**: Перехід на nftables
**Рішення**:

```bash
# Перемикання на legacy iptables
sudo update-alternatives --set iptables /usr/sbin/iptables-legacy
sudo update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy
```

## Автоматичний запуск

### Systemd сервіс (для нативної установки)

Створіть файл `/etc/systemd/system/wg-orbit.service`:

```ini
[Unit]
Description=WireGuard Orbit Server
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/wg-orbit-server run
Restart=always
RestartSec=5
Environment=WG_ORBIT_CONFIG=/etc/wg-orbit/server.yaml

[Install]
WantedBy=multi-user.target
```

```bash
# Увімкнення сервісу
sudo systemctl daemon-reload
sudo systemctl enable wg-orbit
sudo systemctl start wg-orbit
```

### Docker Compose з автозапуском

Створіть файл `docker-compose.yml`:

```yaml
version: '3.8'

services:
  wg-orbit:
    image: ghcr.io/artem/wg-orbit:latest
    container_name: wg-orbit
    restart: unless-stopped
    cap_add:
      - NET_ADMIN
      - SYS_MODULE
    ports:
      - "8080:8080"
      - "51820:51820/udp"
    volumes:
      - /etc/wg-orbit:/etc/wg-orbit
      - /lib/modules:/lib/modules:ro
    environment:
      - WG_ORBIT_CONFIG=/etc/wg-orbit/server.yaml
```

```bash
# Автозапуск з системою
sudo systemctl enable docker
docker-compose up -d
```

## Рекомендації з безпеки

1. **Змініть стандартні порти**:
   ```yaml
   server:
     listen_port: 51821  # Замість 51820
   api:
     port: 8081  # Замість 8080
   ```

2. **Налаштуйте firewall**:
   ```bash
   sudo ufw enable
   sudo ufw allow 22/tcp    # SSH
   sudo ufw allow 8081/tcp  # API (якщо потрібен зовнішній доступ)
   sudo ufw allow 51821/udp # WireGuard
   ```

3. **Регулярно оновлюйте систему**:
   ```bash
   sudo apt update && sudo apt upgrade -y
   sudo reboot  # При необхідності
   ```

## Продуктивність за архітектурами

| Пристрій | Архітектура | Max Peers | Throughput | Рекомендації |
|----------|-------------|-----------|------------|-------------|
| RPi 4 4GB | ARM64 | 100+ | 200+ Mbps | Ідеально для домашнього використання |
| RPi 3 B+ | ARMv7 | 50 | 100 Mbps | Добре для малих мереж |
| Orange Pi 5 | ARM64 | 150+ | 300+ Mbps | Відмінна продуктивність |
| RPi Zero 2 W | ARM64 | 10 | 20 Mbps | Тільки для тестування |

## Підтримка

Якщо у вас виникли проблеми з розгортанням на ARM пристроях:

1. Перевірте [Issues](https://github.com/artem/wg-orbit/issues) на GitHub
2. Створіть новий issue з детальним описом проблеми
3. Включіть інформацію про пристрій: `uname -a`, `cat /proc/cpuinfo`
4. Додайте логи: `docker logs wg-orbit` або `journalctl -u wg-orbit`