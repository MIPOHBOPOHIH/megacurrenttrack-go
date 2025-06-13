# tokenstealer-go

HTTP-прокси на Go, предоставляющий информацию о текущем треке пользователя с сервиса YaMusic через библиотеку [`go-yaynison`](https://github.com/bulatorr/go-yaynison).

## 📦 Описание

Этот сервис запускает HTTP-сервер на `:8080` и предоставляет один эндпоинт:

```
GET /get_current_track_alpha
```

Он использует токен OAuth для подключения к ynison и возвращает информацию о текущем проигрываемом треке.

## 🔐 Аутентификация

Для доступа к эндпоинту требуется заголовок:

```
Authorization: OAuth <your_token>
```

## 📤 Ответы

Успешный ответ:

```json
{
  "paused": true,
  "duration_ms": "130334",
  "progress_ms": "67637",
  "entity_id": "17553811",
  "entity_type": "ARTIST",
  "track_id": "124383437"
}
```

Если очередь пуста:

```json
{
  "error": "PlayerQueue information missing"
}
```

Если токен недействителен:

```json
{
  "error": "Incorrect token entered"
}
```

Если клиент не получил данные от сервера за 10 секунд:

```json
{
  "error": "Failed to retrieve data"
}
```

## 🚀 Запуск

```bash
go build
./main
```

Сервер будет слушать на порту `8080`.

## 🛠️ Зависимости

- [Gin](https://github.com/gin-gonic/gin) — фреймворк для HTTP-сервера.
- [go-yaynison](https://github.com/bulatorr/go-yaynison) — библиотека для работы с WebSocket сервером Ynison.

Установить зависимости можно через:

```bash
go mod download
```

## 📝 Лицензия

MIT — см. [LICENSE](./LICENSE)

---

> Автор: [bulatorr](https://github.com/bulatorr)  
> Репозиторий библиотеки: [go-yaynison](https://github.com/bulatorr/go-yaynison)
