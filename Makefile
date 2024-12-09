include .envrc

.PHONY: run/api
run/api:
	@echo 'Running book club API...'
	@go run ./cmd/api \
	-port=4000 \
	-env=development \
	-db-dsn=${BOOKCLUB_DB_DSN} \
    -smtp-host=${SMTP_HOST} \
    -smtp-port=${SMTP_PORT} \
    -smtp-username=${SMTP_USERNAME} \
    -smtp-password=${SMTP_PASSWORD} \
    -smtp-sender=${SMTP_SENDER} \
    -limiter-rps=2 \
    -limiter-burst=5 \
    -limiter-enabled=false \
    -cors-trusted-origins="http://localhost:9000 http://localhost:9001"

.PHONY: db/psql
db/psql:
	psql ${BOOKCLUB_DB_DSN}

.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations...'
	migrate -path=./migrations -database ${BOOKCLUB_DB_DSN} up