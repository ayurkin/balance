include .env
export $(shell sed 's/=.*//' .env)

run:
	go run cmd/main.go

add_income:
	curl \
	-v \
	--request POST \
	--header "Content-Type: application/json" \
	-d '{"user_id": 1, "value": 10.55, "description": "salary"}' \
	--url http://localhost:3000/balance/v1/income && echo "\n"

add_expense:
	curl \
	-v \
	--request POST \
	--header "Content-Type: application/json" \
	-d '{"user_id": 1, "value": 5.15, "description": "cinema"}' \
	--url http://localhost:3000/balance/v1/expense && echo "\n"

transfer:
	curl \
	-v \
	--request POST \
	--header "Content-Type: application/json" \
	-d '{"user_id_from": 1, "user_id_to": 2, "value": 5, "description": "credit"}' \
	--url http://localhost:3000/balance/v1/transfer && echo "\n"

get_balance:
	curl \
	-v \
	--request GET \
	--url http://localhost:3000/balance/v1/balance?user_id=1 && echo "\n"

tests/integration/balance:
	go test -v ./internal/tests/