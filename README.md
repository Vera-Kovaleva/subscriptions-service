Здравствуйте, это моя реализация проекта по подпискам.

Create = Create
Создает новую подписку.

Read = ReadByID
Получает подписку по её ID.

Update = Update
Обновляет данные существующей подписки.

Delete = Delete
Удаляет подписку по ID.

List = ReadAllByID
Возвращает все подписки пользователя с поддержкой пагинации.

TotalCost = TotalCost 
Вычисляет общую стоимость подписок за указанный период. Если end не указан, используется текущая дата.

Примеры использования
Создание подписки
POST /subscriptions
Content-Type: application/json

{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "service_name": "Music",
  "price": 1599,
  "start_date": "01-2024",
  "end_date": "12-2024"
}

Получение подписки по ID
GET /subscriptions/{id}

Получение всех подписок пользователя
GET /subscriptions?user_id={user_id}&limit=50&offset=0

Вычисление общей стоимости подписок
GET /subscriptions/total?user_id={user_id}&service_name=Music&start_date=01-2024&end_date=12-2024