.PHONY: run migrate seed test lint build

build:
	go build -o bin/dormitory-bot ./cmd/dormitory-bot/main.go

run:
	go run ./cmd/dormitory-bot/main.go

migrate:
	SEED_MODE=migrate go run ./cmd/dormitory-bot/main.go

seed:
	SEED_MODE=seed go run ./cmd/dormitory-bot/main.go

test:
	go test ./... -v -count=1

lint:
	go vet ./...

mod-tidy:
	go mod tidy
