# Микросервис для работы с балансом пользователей

**Проблема:**

В нашей компании есть много различных микросервисов. Многие из них так или иначе хотят взаимодействовать с балансом пользователя. На архитектурном комитете приняли решение централизовать работу с балансом пользователя в отдельный сервис.

**Задача:**

Необходимо реализовать микросервис для работы с балансом пользователей (зачисление средств, списание средств, перевод средств от пользователя к пользователю, а также метод получения баланса пользователя). Сервис должен предоставлять HTTP API и принимать/отдавать запросы/ответы в формате JSON.

**Сценарии использования:**

Далее описаны несколько упрощенных кейсов приближенных к реальности.
1. Сервис биллинга с помощью внешних мерчантов (аля через visa/mastercard) обработал зачисление денег на наш счет. Теперь биллингу нужно добавить эти деньги на баланс пользователя.
2. Пользователь хочет купить у нас какую-то услугу. Для этого у нас есть специальный сервис управления услугами, который перед применением услуги проверяет баланс и потом списывает необходимую сумму.
3. В ближайшем будущем планируется дать пользователям возможность перечислять деньги друг-другу внутри нашей платформы. Мы решили заранее предусмотреть такую возможность и заложить ее в архитектуру нашего сервиса.

## Запуск и работа микросервиса

### Запуск микросервиса и БД Postgres

```
make run
```

### Эндпоинты

**Метод начисления средств на баланс. Принимает id пользователя, сколько средств зачислить, описание операции**

```
curl \
-v \
--request POST \
--header "Content-Type: application/json" \
-d '{"user_id": 1, "value": 10.55, "description": "salary"}' \
--url http://localhost:3000/balance/v1/income && echo "\n"

или

make add_income
```

Ответ

```
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Wed, 05 Oct 2022 18:02:25 GMT
{"status": "success"}
```

**Метод списания средств с баланса. Принимает id пользователя, сколько средств списать, описание операции**

```
curl \
-v \
--request POST \
--header "Content-Type: application/json" \
-d '{"user_id": 1, "value": 5.15, "description": "cinema"}' \
--url http://localhost:3000/balance/v1/expense && echo "\n"

или

make add_expense
```

Ответ

```
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Wed, 05 Oct 2022 18:02:30 GMT
{"status": "success"}
```
Если списать средства у несуществующего пользователя
```
< HTTP/1.1 400 Bad Request
< Content-Type: application/json
< Date: Wed, 05 Oct 2022 18:05:52 GMT
{"errorText": "user_id 10: user_id does not exist"}
```
Если списать средств больше, чем есть у пользователя
```
< HTTP/1.1 400 Bad Request
< Content-Type: application/json
< Date: Wed, 05 Oct 2022 18:05:52 GMT
{"errorText": "user_id 1: user_id has not enough balance"}
```

**Метод перевода средств от пользователя к пользователю. Принимает id пользователя с которого нужно списать средства, id пользователя которому должны зачислить средства, а также сумму и описание операции**

```
curl \
-v \
--request POST \
--header "Content-Type: application/json" \
-d '{"user_id_from": 1, "user_id_to": 2, "value": 5, "description": "credit"}' \
--url http://localhost:3000/balance/v1/transfer && echo "\n"

или 

make transfer
```

Ответ

```
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Wed, 05 Oct 2022 18:07:30 GMT
{"status": "success"}
```
Если перевести средства от несуществующего пользователя
```
< HTTP/1.1 400 Bad Request
< Content-Type: application/json
< Date: Wed, 05 Oct 2022 18:08:52 GMT
{"errorText": "user_id 10: user_id does not exist"}
```
Если перевести средств больше, чем есть у пользователя
```
< HTTP/1.1 400 Bad Request
< Content-Type: application/json
< Date: Wed, 05 Oct 2022 18:09:10 GMT
{"errorText": "user_id 1: user_id has not enough balance"}
```

**Метод получения текущего баланса пользователя. Принимает id пользователя. Баланс всегда в рублях**

```
curl \
-v \
--request GET \
--url http://localhost:3000/balance/v1/balance?user_id=1 && echo "\n"

или

make get_balance
```

Ответ

```
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Wed, 05 Oct 2022 18:12:30 GMT
{"user_id":1,"value":"16.35"}
```
Если получить баланс у несуществующего пользователя
```
< HTTP/1.1 400 Bad Request
< Content-Type: application/json
< Date: Wed, 05 Oct 2022 18:22:52 GMT
{"errorText": "user_id 10: user_id does not exist"}
```

## Запуск интеграционных тестов

```
go test -v ./internal/tests/

или

make tests/integration/balance
```
