# Тестовое задание на стажировку Avito 2024

## Описание
Сервис, который позволяет показывать пользователям баннеры, в зависимости от требуемой фичи и тега пользователя, а также управлять баннерами и связанными с ними тегами и фичами.

Сущности БД находятся в папке `postgres`.

Коллекция интеграционных тестов и коллекция постмана для нагрузочного тестирования находятся в папке `postman`.

При добавлении 1000 тегов, фич и баннеров. Время выполнения запроса на получение банера для пользователя, по данным из постмана, была около 4 мс.


## Стек
1. Golang
2. Postgres
3. Gin
4. Docker
5. go-cache

## Quickstart
1. 💻 Скачай проект
2. ✅ Запусти проект: `docker compose up`

## Credentials
Токен отправляется в Header `token`

- Токен администратора: `admin`
- Токен пользователя: `user`

## Примеры использования

### POST Добавить баннер
`http://localhost:8080/banner`
#### Body
```
{
  "tag_ids": [
    1,2,7
  ],
  "feature_id": 4,
  "content": {
    "title": "Alice",
    "text": "alice",
    "url": "alice.com"
  },
  "is_active": true
}
```

#### Response
```
{
    "banner_id": 1
}
```

### GET Получить баннер
`http://localhost:8080/user_banner?tag_id=1&feature_id=4`

#### Response
```
{
    "text": "alice",
    "title": "Alice",
    "url": "alice.com"
}
```

### PATCH Изменить баннер
`http://localhost:8080/banner/1`
#### Body
```
{
  "tag_ids": [
    1,2,9
  ],
  "feature_id": 5,
  "content": {
    "title": "Alice2",
    "text": "alice2",
    "url": "alice2.com"
  },
  "is_active": false
}
```

### GET Получить список баннеров
`http://localhost:8080/user_banner?tag_id=1&feature_id=4`

#### Response
```
[
    {
        "banner_id": 1,
        "tag_ids": [
            1,
            2,
            9
        ],
        "feature_id": 5,
        "content": {
            "text": "alice2",
            "title": "Alice2",
            "url": "alice2.com"
        },
        "is_active": false,
        "created_at": "2024-04-14T19:53:26Z",
        "updated_at": "2024-04-14T19:53:44Z"
    }
]
```

### DELETE Удалить баннер
`http://localhost:8080/banner/1`