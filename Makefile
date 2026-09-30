include .env
export

.PHONY: db-up db-down migrate-up migrate-down run

db-up:
	docker compose up -d

db-down:
	docker compose down

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

run:
	go run ./cmd/api