# URL Shortener

## Команды Makefile

```sh
make up-postgres   # Сборка и запуск с PostgreSQL и миграциями
make up-inmemory   # Сборка и запуск с хранением в памяти
make down          # Остановка и удаление контейнеров, без удаления volume БД
make test          # Все тесты с детектором гонок
make test-service  # Тесты сервиса с подробным выводом
make test-handler  # Тесты HTTP-хендлеров с подробным выводом
```

При запуске выбранного режима текущие контейнеры проекта останавливаются.
Данные PostgreSQL сохраняются в volume; данные in-memory теряются.

## Запуск с PostgreSQL

```sh
docker compose up --build -d
```

Compose запускает PostgreSQL, применяет миграции и запускает API на порту 8080.

## Запуск с хранением в памяти

```sh
docker compose -f docker-compose.inmemory.yaml up --build -d
```

Это отдельная конфигурация без PostgreSQL и миграций. Не запускайте оба
варианта одновременно на одном порту. Данные в памяти теряются при перезапуске.

## Выбор хранилища при запуске приложения

Параметр `-storage` принимает `postgres` или `inmemory` и имеет приоритет
над переменной `STORAGE`. По умолчанию используется `postgres`.

```sh
BASE_URL=http://localhost:8080 APP_PORT=8080 go run ./cmd/server -storage=inmemory
```

Для `postgres` также требуется `DATABASE_URL` и предварительное применение
миграций из каталога `migrations`.

`BASE_URL` задаёт публичный адрес для выдаваемых ссылок. В Compose внешний
порт задаёт `APP_PORT` (по умолчанию 8080); при отсутствии `BASE_URL`
используется `http://localhost:<APP_PORT>`.

## API

```sh
curl -X POST http://localhost:8080/api/v1/urls \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://ozon.ru"}'

curl --get http://localhost:8080/api/v1/urls \
  --data-urlencode 'short_url=http://localhost:8080/XXXXXXXXXX'
```

Замените `XXXXXXXXXX` на код из ответа POST. Оригинальный URL ограничен
2000 байтами; короткий код состоит из 10 латинских букв, цифр или `_`.
