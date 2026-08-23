.PHONY: fmt test race vet web-test web-build build verify migrate-up seed

fmt:
	gofmt -w cmd internal tests

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

web-test:
	cd web && npm test

web-build:
	cd web && npm run build

build:
	go build ./cmd/server

verify: fmt test race vet web-test web-build build

migrate-up:
	psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f migrations/000001_initial.up.sql

seed:
	psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed.sql
