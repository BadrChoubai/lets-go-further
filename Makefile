include .envrc

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api \
	-db-dsn=${GREENLIGHT_DB_DSN} \
	-smtp-USERNAME=${SMTP_USERNAME} \
	-smtp-password=${SMTP_PASSWORD}

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${GREENLIGHT_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up

## db/migrations/refresh: apply all down and up database migrations
.PHONY: db/migrations/refresh
db/migrations/refresh:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} down
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up

## audit: tidy dependencies and format, vet, and test all code
.PHONY: audit
audit:
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies'
	go mod vendor

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building: cmd/api...'
	go build -ldflags='-s' -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api

production_host_ip = 'xxx.xxx.xxx.xxx'

## production/connect: connect to the production server
.PHONY: production/connect
production/connect:
	ssh greenlight@${production_host_ip}

## production/ping: ping healthcheck endpoint
.PHONY: production/ping
production/ping:
	curl -Li http://${production_host_ip}/healthcheck

## production/debug/snapshot: create auto-closing ssh tunnel, curl /debug/vars and save its output to file
.PHONY: production/debug/snapshot
production/debug/snapshot:
	mkdir -p ./remote/debug
	ssh -f -L :9999:${production_host_ip}:4000 greenlight@${production_host_ip} sleep 5; \
	curl -o ./remote/debug/snapshot.json localhost:9999/debug/vars

## production/deploy/api: request /debug/vars and save output to file
.PHONY: production/deploy/api
production/deploy/api:
	rsync -P ./bin/linux_amd64/api greenlight@${production_host_ip}:~
	rsync -rP --delete ./migrations greenlight@${production_host_ip}:~
	rsync -P ./remote/production/api.service greenlight@${production_host_ip}:~
	rsync -P ./remote/production/Caddyfile greenlight@${production_host_ip}:~
	ssh -t greenlight@${production_host_ip} '\
		migrate -path ~/migrations -database $$GREENLIGHT_DB_DSN up \
		&& sudo mv ~/api.service /etc/systemd/system/ \
		&& sudo systemctl enable api \
		&& sudo systemctl restart api \
		&& sudo mv ~/Caddyfile /etc/caddy/ \
		&& sudo systemctl reload caddy \
	'

