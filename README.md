# MEGACURRENTTRACK-GO

МЕГАКРУТОЙ ФОРК [`ОЧЕНЬ КРУТОГО ПРОЕКТА`](https://github.com/bulatorr/tokenstealer-go), предоставляющий информацию о текущем треке пользователя с сервиса YaMusic через библиотеку [`go-yaynison`](https://github.com/bulatorr/go-yaynison).

## 📦 Описание

Этот сервис запускает HTTP-сервер на `:8080` и предоставляет один эндпоинт:

```
GET /get_current_track_beta // YaMusic
```

Он использует токен OAuth для подключения к ynison и возвращает информацию о текущем проигрываемом треке.

## 🔐 Аутентификация

Для доступа к эндпоинту /get_current_track_beta требуется заголовок:

```
ya-token: token // https://github.com/MarshalX/yandex-music-api/discussions/513
```


## 📤 Ответы /get_current_track_beta

Успешный ответ:

```json
{
    "duration_ms": "170010",
    "progress_ms": "53361",
    "entity_id": "947609922:3",
    "entity_type": "PLAYLIST",
    "track": {
        "track_id": 130839122,
        "title": "ЭГО ТРИП II",
        "artist": "dekma",
        "img": "https://avatars.yandex.net/get-music-content/14299670/0e8ba055.a.33173515-1/1000x1000",
        "duration": 170,
        "minutes": 2,
        "seconds": 50,
        "album_id": 33173515,
        "download_link": "тут я убрал ссылку думаете понимаю почему"
    }
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

> Автор оригинала: [bulatorr](https://github.com/bulatorr)  
> Репозиторий библиотеки: [go-yaynison](https://github.com/bulatorr/go-yaynison)
> Сделал получение ссылки на скачивание: [atennop](https://atennop.tech)
