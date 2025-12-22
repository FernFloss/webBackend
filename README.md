# Развертывание Backend (Docker Compose)

## Предусловия

### 1. Создание Docker сети (если еще не создана)
```bash
docker network create tyk-network
```

### 2. Запущенные зависимости
Убедитесь, что запущены и доступны:
- **PostgreSQL** (порт 5432)
- **RabbitMQ** (порт 5672)

Оба сервиса должны быть в сети `tyk-network` или доступны по указанным адресам.

## Запуск

### 1. Создание файла конфигурации
```bash
cp env.example .env
```

### 2. Настройка .env файла
Отредактируйте `.env` файл в соответствии с вашей инфраструктурой:

**Важно:** В `env.example` есть несоответствие в комментариях:
- Комментарий говорит использовать `localhost` для RabbitMQ
- Но в `docker-compose.yaml` используется имя контейнера `rabbitmq`

**Если PostgreSQL и RabbitMQ запущены в Docker контейнерах в сети `tyk-network`**, используйте имена контейнеров:
```env
DB_HOST=postgres
RABBITMQ_URL=amqp://rabbitmq_user:rabbitmq_pass@rabbitmq:5672/
```

**Если PostgreSQL и RabbitMQ запущены на внешних хостах**, используйте IP-адреса или доменные имена:
```env
DB_HOST=172.18.130.222  # или localhost
RABBITMQ_URL=amqp://rabbitmq_user:rabbitmq_pass@172.18.130.222:5672/
```

### 3. Запуск сервиса

**Первый запуск или после изменений в коде:**
```bash
docker-compose up -d --build
```

**Обычный запуск:**
```bash
docker-compose up -d
```

## Проверка

### Проверка health endpoint
```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:
```json
{"status":"ok"}
```

### Просмотр логов
```bash
# Все сервисы
docker-compose logs -f

# Только backend сервис
docker-compose logs -f report-service
```

### Проверка статуса
```bash
docker-compose ps
```

### Проверка подключения к БД
Проверьте логи на наличие сообщений:
- `Successfully connected to database` - успешное подключение
- `Failed to connect to database` - ошибка подключения

### Проверка подключения к RabbitMQ
Проверьте логи на наличие сообщений:
- `Connected to RabbitMQ` - успешное подключение
- `Failed to initialize RabbitMQ` - ошибка подключения

## Остановка
```bash
docker-compose down
```
## Устранение неполадок

### Сервис не запускается
- Проверьте логи: `docker-compose logs report-service`
- Убедитесь, что порт 8080 не занят другим приложением
- Проверьте правильность настроек в `.env` файле

### Ошибка подключения к PostgreSQL
- Убедитесь, что PostgreSQL запущен и доступен
- Проверьте правильность `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD` в `.env`
- Убедитесь, что PostgreSQL находится в той же сети `tyk-network` или доступен по указанному адресу
- Проверьте, что база данных `auditorium_db` существует (миграции создадут её автоматически)

### Ошибка подключения к RabbitMQ
- Убедитесь, что RabbitMQ запущен и доступен
- Проверьте правильность `RABBITMQ_URL` в `.env` файле
- Убедитесь, что RabbitMQ находится в той же сети `tyk-network` или доступен по указанному адресу
- Проверьте учетные данные в URL (username:password)

### Health endpoint не отвечает
- Проверьте, что контейнер запущен: `docker-compose ps`
- Проверьте логи на наличие ошибок: `docker-compose logs report-service`
- Убедитесь, что порт 8080 проброшен корректно: `docker-compose port report-service 8080`

### Проверка подключения к сети
```bash
docker network inspect tyk-network
```

## Разработчики

Садковская Маргарита, Милованова Анастасия, Пивоварова Милена

## Ссылки

Frontend-часть -- https://github.com/TheSleepySparrow/frontend_track_occupancy
ML-часть -- https://github.com/FernFloss/webML