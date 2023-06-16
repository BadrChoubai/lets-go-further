PRIMARY_DB = ${GREENLIGHT_DB_DSN}

.PHONY: run
run:
	go run ./cmd/api

.PHONY: migrate_up
migrate_up: migrate_down
	migrate -path ./migrations -database ${PRIMARY_DB} up

.PHONY: migrate_down
migrate_down:
	migrate -path ./migrations -database ${PRIMARY_DB} down
